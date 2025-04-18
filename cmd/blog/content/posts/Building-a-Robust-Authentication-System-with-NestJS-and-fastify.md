---
title: "Building a Robust Authentication System with NestJS"
date: 2024-8-27
excerpt: " NestJS: Building a Robust Authentication System with NestJS"
---
# Building a Robust Authentication System with NestJS

## Introduction
---
In this comprehensive blog post, I'll walk you through the process of building a robust authentication system using NestJS, a powerful Node.js framework. We'll cover essential aspects of user management, authentication strategies, and secure API development. This project encompasses user registration, credential validation, JWT token generation, and role-based authentication for both users and admins.

## Setting Up the UsersService
---
The `UsersService` is responsible for handling user-related operations, such as user registration and retrieval of all registered users.

### Dependencies

The `UsersService` has the following dependency:

```typescript
@InjectModel(User.name) private userModel: Model<UserDocument>
```

This injects the Mongoose model for the `User` schema, allowing the service to interact with the MongoDB database.

### User Registration

The `registerUser` method handles new user registration:

```typescript
async registerUser(registerUserDto: RegisterUserDto): Promise<User> {
  const { userName, email, password } = registerUserDto;

  const existingUser = await this.userModel.findOne({ email });
  if (existingUser) {
    throw new ConflictException('Email already in use');
  }

  const salt = await bcrypt.genSalt(14);
  const hashedPassword = await bcrypt.hash(password, salt);

  const newUser = new this.userModel({
    userName,
    email,
    password: hashedPassword,
  });

  return newUser.save();
}
```

This method:
1. Extracts user information from the DTO.
2. Checks if a user with the provided email already exists.
3. If the email is unique, it generates a salt and hashes the password.
4. Creates and saves a new `User` document in the database.

### User Retrieval

The `findAll` method retrieves all registered users:

```typescript
async findAll(): Promise<User[]> {
  return this.userModel.find().exec();
}
```

This method simply returns all `User` documents from the database.

## Configuring Nginx as a Reverse Proxy
---
To enhance our application's scalability and security, we set up an Nginx server to act as a reverse proxy. Here's the configuration:

```nginx
server {
    listen 80;
    server_name localhost;

    location / {
        proxy_pass http://127.0.0.1:3000/users;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    error_page 404 /404.html;
    location = /404.html {
        root /usr/share/nginx/html;
        internal;
    }

    error_page 500 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
```

This configuration:
- Listens on port 80
- Proxies requests to our NestJS app running on `http://127.0.0.1:3000/users`
- Handles WebSocket upgrades
- Sets appropriate headers for proxy communication
- Configures error pages for 404 and 50x errors

## Implementing Authentication Strategies
---
We implemented two authentication strategies: one for regular users and another for admin users.

### LocalUserStrategy

```typescript
@Injectable()
export class LocalUserStrategy extends PassportStrategy(Strategy, 'local-user') {
  constructor(private authService: AuthService) {
    super({
      usernameField: 'email',
      passwordField: 'password',
    });
  }

  async validate(email: string, password: string): Promise<any> {
    console.log(`LocalUserStrategy validating: ${email}`);
    const user = await this.authService.validateUserCredentials(email, password);
    if (!user) {
      console.log(`User validation failed for email: ${email}`);
      throw new UnauthorizedException('Invalid email or password');
    }
    console.log(`User validation successful for email: ${email}`);
    return user;
  }
}
```

### LocalAdminStrategy

```typescript
@Injectable()
export class LocalAdminStrategy extends PassportStrategy(Strategy, 'local-admin') {
  constructor(private readonly authService: AuthService) {
    super({
      usernameField: 'email',
      passwordField: 'password',
    });
  }

  async validate(email: string, password: string): Promise<any> {
    console.log(`LocalAdminStrategy validating: ${email}`);
    const admin = await this.authService.validateAdminCredentials(email, password);
    if (!admin) {
      console.log(`Admin validation failed for email: ${email}`);
      throw new UnauthorizedException('Invalid email or password for admin');
    }
    console.log(`Admin validation successful for email: ${email}`);
    return admin;
  }
}
```

Both strategies use Passport.js and extend `PassportStrategy`. They differ in the strategy name ('local-user' vs 'local-admin') and the validation method they call from the `AuthService`.

## AuthService Implementation
---
The `AuthService` manages user and admin authentication. Here are the key methods:

### Validating User Credentials

```typescript
async validateUserCredentials(email: string, password: string): Promise<UserDocument | null> {
  console.log(`Validating user credentials for email: ${email}`);
  const user = await this.userModel.findOne({ email, role: 'User' });
  if (!user) {
    console.log(`User not found for email: ${email}`);
    return null;
  }
  const isPasswordValid = await bcrypt.compare(password, user.password);
  console.log(`Password validation result: ${isPasswordValid}`);
  return isPasswordValid ? user : null;
}
```

### Validating Admin Credentials

```typescript
async validateAdminCredentials(email: string, password: string): Promise<UserDocument | null> {
  console.log(`Validating admin credentials for email: ${email}`);
  const admin = await this.userModel.findOne({ email, role: 'Admin' });
  if (!admin) {
    console.log(`Admin not found for email: ${email}`);
    return null;
  }
  const isPasswordValid = await bcrypt.compare(password, admin.password);
  console.log(`Password validation result: ${isPasswordValid}`);
  return isPasswordValid ? admin : null;
}
```

### Generating JWT Token for Users

```typescript
async loginUser(user: UserDocument): Promise<any> {
  console.log(`User login successful for email: ${user.email}`);
  const payload = { email: user.email, sub: user._id, role: user.role };
  const accessToken = this.jwtService.sign(payload, {
    secret: this.configService.get<string>('JWT_SECRET_USER'),
  });
  const { password, ...result } = user.toJSON();
  return { ...result, accessToken };
}
```

### Generating JWT Token for Admins

```typescript
async loginAdmin(admin: UserDocument): Promise<any> {
  console.log(`Admin login successful for email: ${admin.email}`);
  const payload = { email: admin.email, sub: admin._id, role: admin.role };
  const accessToken = this.jwtService.sign(payload, {
    secret: this.configService.get<string>('JWT_SECRET_ADMIN'),
  });
  const { password, ...result } = admin.toJSON();
  return { ...result, accessToken };
}
```

Note that we use separate JWT secrets for users and admins to enhance security.

## AuthController Setup

The `AuthController` handles authentication endpoints for logging in users and admins:

```typescript
@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @UseGuards(AuthGuard('local-user'))
  @Post('login-user')
  async loginUser(@Req() req) {
    console.log(req.user);
    return this.authService.loginUser(req.user);
  }

  @UseGuards(AuthGuard('local-admin'))
  @Post('login-admin')
  async loginAdmin(@Req() req) {
    console.log(req.user);
    return this.authService.loginAdmin(req.user);
  }
}
```

This controller uses Guards to ensure proper authentication before processing login requests for both users and admins.

## Key Takeaways

1. **Modular Design**: NestJS's modular architecture allowed us to separate concerns effectively, making the codebase more maintainable.

2. **Security First**: By implementing bcrypt for password hashing and JWT for token generation, we prioritized security in our authentication system.

3. **Scalability**: The Nginx reverse proxy setup prepares our application for potential scaling needs.

4. **Role-Based Authentication**: By distinguishing between user and admin authentication, we've laid the groundwork for role-based access control.

5. **Logging for Debugging**: We've implemented extensive logging throughout the authentication process, which will be invaluable for debugging and monitoring.



## Main Application Setup
---
In our project, we've designed the main application setup to be flexible, allowing us to easily switch between different HTTP providers such as Express (the default) and Fastify. This approach gives us the freedom to choose the most suitable provider based on our needs without major code changes. Let's look at how we set up our main NestJS application:

```typescript
import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);
  await app.listen(3000, '0.0.0.0');
}
bootstrap();
```

Let's break down this code:

1. We import the necessary modules:
   - `NestFactory` from `@nestjs/core` to create our NestJS application
   - `AppModule` from './app.module', which is our root module

2. We define an asynchronous `bootstrap` function:
   - We create a new NestJS application using `NestFactory.create`
   - We pass `AppModule` as the root module of our application

3. We call `app.listen(3000, '0.0.0.0')` to start our server:
   - The server listens on port 3000
   - '0.0.0.0' means it will listen on all available network interfaces, not just localhost

4. Finally, we call the `bootstrap` function to start our application

This setup allows us to use NestJS's default HTTP provider (Express) out of the box. However, if we want to switch to Fastify or any other supported HTTP provider, we can do so with minimal changes:

1. Install the necessary package (e.g., `@nestjs/platform-fastify` for Fastify)
2. Import the appropriate adapter (e.g., `FastifyAdapter`)
3. Modify the `NestFactory.create()` call to use the new adapter

For example, to use Fastify, we would modify our code like this:

```typescript
import { NestFactory } from '@nestjs/core';
import { FastifyAdapter, NestFastifyApplication } from '@nestjs/platform-fastify';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create<NestFastifyApplication>(
    AppModule,
    new FastifyAdapter()
  );
  await app.listen(3000, '0.0.0.0');
}
bootstrap();
```

This flexible approach allows us to:

1. Start with Express (NestJS's default) for rapid development and wide compatibility
2. Easily switch to Fastify if we need its high-performance capabilities
3. Adapt to other HTTP providers as our project requirements evolve


## Conclusion

Building a secure and efficient authentication system is crucial for any web application. Through this project, we've created a solid foundation using NestJS, integrating best practices in user management, authentication, and API security.

We've covered:
- Setting up a `UsersService` for user registration and retrieval
- Configuring Nginx as a reverse proxy
- Implementing separate authentication strategies for users and admins
- Creating an `AuthService` to handle credential validation and JWT token generation
- Setting up an `AuthController` to expose login endpoints

---
