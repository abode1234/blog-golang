---
title: "Organizing a Video Streaming Application in Rust: File-by-File Breakdown"
date: 2024-06-30
excerpt: "Organizing a Video Streaming Application in Rust: File-by-File Breakdown"
---

## Organizing a Video Streaming Application in Rust: File-by-File Breakdown

### Introduction

When developing a video streaming application, modularizing your code into well-defined components is crucial for maintainability and scalability. In this article, we will break down the structure of a Rust project that streams video using SDL2, dividing the logic into separate files for clarity and reusability.

### Project Structure

The project is organized into several Rust source files within the `src/` directory, each handling a specific part of the video streaming process:

```
videos-stream/
│
├── src/
│   ├── frame_loader.rs        // Responsible for loading image frames
│   ├── video_stream.rs        // Manages the stream of video frames
│   ├── video_renderer.rs      // Handles the rendering of video frames
│   └── main.rs                // Entry point of the application
│
├── frames/                    // Directory containing the image frames (not included in the repository)
├── Cargo.toml                 // Cargo configuration file
└── .gitignore                 // To exclude unnecessary files and directories
```

Let’s delve into each file to understand its role.

### `frame_loader.rs` - Loading Image Frames

The `frame_loader.rs` file is responsible for reading image files from the `frames/` directory and converting them into SDL2 textures. This modular approach allows for easy changes if, for example, the source of frames changes from disk files to network streams or another input method.

```rust
// src/frame_loader.rs

use sdl2::image::LoadTexture;
use sdl2::render::Texture;
use std::collections::VecDeque;
use std::fs;

pub fn load_frames<'a>(
    texture_creator: &'a sdl2::render::TextureCreator<sdl2::video::WindowContext>,
    frames_dir: &str,
) -> VecDeque<Texture<'a>> {
    let frames: Vec<_> = fs::read_dir(frames_dir)
        .expect("Failed to read frames directory")
        .filter_map(|entry| entry.ok())
        .map(|entry| entry.path())
        .filter(|path| path.is_file())
        .collect();

    let mut textures = VecDeque::new();
    for frame_path in frames {
        let texture = texture_creator
            .load_texture(&frame_path)
            .expect("Failed to load texture from image");
        textures.push_back(texture);
    }

    textures
}
```

### `video_stream.rs` - Managing the Video Stream

The `video_stream.rs` file defines the `VideoStream` struct, which manages the queue of textures that represent the video frames. It provides methods to retrieve the next frame in the stream, maintaining the sequence and flow of the video.

```rust
// src/video_stream.rs

use sdl2::render::Texture;
use std::collections::VecDeque;

pub struct VideoStream<'a> {
    textures: VecDeque<Texture<'a>>,
}

impl<'a> VideoStream<'a> {
    pub fn new(textures: VecDeque<Texture<'a>>) -> VideoStream<'a> {
        VideoStream { textures }
    }

    pub fn next_frame(&mut self) -> Option<Texture<'a>> {
        self.textures.pop_front()
    }
}
```

### `video_renderer.rs` - Rendering the Video Frames

The `video_renderer.rs` file focuses on the rendering logic. It uses SDL2’s canvas to draw the textures (frames) onto the window. It also manages timing to ensure that the video plays at the correct frames per second (FPS).

```rust
// src/video_renderer.rs

use sdl2::Sdl;
use std::time::{Duration, Instant};

use crate::video_stream::VideoStream;

pub fn render_video_at_fps(sdl_context: &Sdl, mut video_stream: VideoStream, target_fps: u32) {
    let video_subsystem = sdl_context.video().unwrap();
    let window = video_subsystem
        .window("Video Stream", 800, 600)
        .position_centered()
        .build()
        .expect("Failed to create window");

    let mut canvas = window
        .into_canvas()
        .present_vsync()
        .build()
        .expect("Failed to create canvas");

    let target_frame_duration = Duration::from_secs_f64(1.0 / target_fps as f64);

    let mut frame_count = 0;
    let mut start_time = Instant::now();

    while let Some(texture) = video_stream.next_frame() {
        let frame_start = Instant::now();

        canvas.clear();
        canvas.copy(&texture, None, None).expect("Failed to copy texture to canvas");
        canvas.present();

        frame_count += 1;

        let elapsed_time = frame_start.duration_since(start_time).as_secs_f64();
        if elapsed_time >= 1.0 {
            let average_fps = frame_count as f64 / elapsed_time;
            println!("Frames per second: {:.2}", average_fps);

            frame_count = 0;
            start_time = Instant::now();
        }

        let frame_duration = frame_start.elapsed();
        if let Some(sleep_duration) = target_frame_duration.checked_sub(frame_duration) {
            std::thread::sleep(sleep_duration);
        }
    }
}
```

### `main.rs` - Application Entry Point

Finally, the `main.rs` file ties everything together. It initializes SDL2, loads the video frames using `frame_loader.rs`, creates a `VideoStream` object, and then calls the renderer to start the video stream.

```rust
// src/main.rs

mod frame_loader;
mod video_stream;
mod video_renderer;

fn main() {
    let sdl_context = sdl2::init().expect("Failed to initialize SDL2");

    let frames_dir = "frames"; // Directory containing the image frames
    let target_fps = 90;

    let texture_creator = sdl_context
        .video()
        .unwrap()
        .window("Video Stream", 800, 600)
        .position_centered()
        .build()
        .expect("Failed to create window")
        .into_canvas()
        .present_vsync()
        .build()
        .expect("Failed to create canvas")
        .texture_creator();

    let textures = frame_loader::load_frames(&texture_creator, frames_dir);
    let video_stream = video_stream::VideoStream::new(textures);

    video_renderer::render_video_at_fps(&sdl_context, video_stream, target_fps);
}
```

### Conclusion

By breaking down the video streaming application into these modular components, we've created a clean, maintainable codebase that can be easily extended or modified. Each file has a clear responsibility, from loading frames to managing the video stream and rendering frames at a consistent FPS.

This structure not only makes the code easier to understand but also facilitates future enhancements, such as adding new features or optimizing performance. By following this approach, you can apply similar principles to other Rust projects, ensuring they remain scalable and maintainable over time.

---

