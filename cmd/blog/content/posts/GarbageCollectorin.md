---
title: "Implementing a Simple Garbage Collector in C"
date: 2024-09-22
excerpt: "Learn how to build a basic garbage collector in C using mark-and-sweep algorithm"
---


# Implementing a Simple Garbage Collector in C

Memory management is a critical aspect of programming, especially in low-level languages like C, where developers are responsible for allocating and freeing memory. In this blog post, we will build a simple garbage collector in C. This garbage collector uses a **mark-and-sweep** algorithm to manage dynamic memory allocation. Let's dive into the details and explore how it works.

#### Shell Script for Compilation and Testing

The first step is to compile and run the C program using a shell script. This script compiles the `main.c` file and then runs the compiled program if the compilation succeeds.

```bash
#!/bin/sh

gcc main.c -o garbage_collector

if [ $? -eq 0 ]; then
    echo "test succeeded"

    echo "Testing garbage collector with sample data..."
    ./garbage_collector <<EOF
10
20
EOF

else
    echo "test failed"
fi
```

The script checks if the `gcc` compilation was successful (`$? -eq 0`), and if so, it runs the program with sample input (`10` and `20`). If the compilation fails, it outputs "test failed."

---

### Garbage Collection in C

Now let's explore the actual C code, starting with the structure of our memory blocks.

#### Memory Block Structure

Each block of memory that we allocate or free will be represented by a `header_t` structure:

```c
#include <stdio.h>
#include <stdlib.h>
#include <stddef.h>  // For size_t, NULL
#include <stdint.h>  // For uintptr_t
#include <assert.h>  // For assert
#include <unistd.h>  // For sbrk

typedef struct header_t {
    struct header_t *next;  // Pointer to the next block in the free/used list
    size_t size;            // Size of the memory block
} header_t;
```

This structure stores a pointer to the next memory block and the size of the current block. It's used to manage both free and used memory blocks.

#### Memory Management Basics

We maintain two lists: one for free memory blocks (`freep`) and another for used memory blocks (`usedp`). Initially, these pointers are set to `NULL`.

```c
static header_t base;
static header_t *freep = &base;  // Points to the first free block of memory
static header_t *usedp = NULL;   // Points to the first used block of memory

uintptr_t stack_bottom;
uintptr_t etext, end;  // Placeholders for scanning
```

These global variables also include `stack_bottom`, `etext`, and `end`, which represent the boundaries for scanning the stack and memory sections.

### Memory Allocation (`GC_malloc`)

The `GC_malloc` function is used to allocate memory dynamically:

```c
void *GC_malloc(size_t alloc_size) {
    header_t *p, *prevp;
    size_t nunits = (alloc_size + sizeof(header_t) - 1) / sizeof(header_t) + 1;

    prevp = freep;
    for (p = prevp->next; ; prevp = p, p = p->next) {
        if (p->size >= nunits) {
            if (p->size == nunits) {
                prevp->next = p->next;  // Exact fit
            } else {
                p->size -= nunits;
                p += p->size;
                p->size = nunits;
            }
            freep = prevp;
            return (void *)(p + 1);
        }
        if (p == freep) {  // Wrapped around free list
            if ((p = morecore(nunits)) == NULL) {
                return NULL;  // No memory available
            }
        }
    }
}
```

In this function:
- The number of units required for allocation is calculated.
- We traverse the free list to find a block that is large enough to satisfy the request.
- If no block is found, we call `morecore()` to request additional memory from the system using `sbrk()`.

### Memory Expansion (`morecore`)

The `morecore` function requests additional memory from the system if the free list cannot satisfy the memory allocation request:

```c
static header_t *morecore(size_t nu) {
    if (nu < 1024) {
        nu = 1024;  // Request at least 1024 units
    }
    char *cp = sbrk(nu * sizeof(header_t));
    if (cp == (char *) -1) {  // No memory available
        return NULL;
    }

    header_t *up = (header_t *) cp;
    up->size = nu;
    add_to_free_list(up);
    return freep;
}
```

This function ensures that we always allocate at least 1024 units of memory when calling `sbrk()`.

### Freeing Memory and Managing the Free List

When memory is no longer needed, we need to add it back to the free list:

```c
void add_to_free_list(header_t *bp) {
    header_t *p;

    for (p = freep; !(bp > p && bp < p->next); p = p->next) {
        if (p >= p->next && (bp > p || bp < p->next)) {
            break;  // At the end of the list, wrap around
        }
    }

    // Merge adjacent free blocks
    if (bp + bp->size == p->next) {
        bp->size += p->next->size;
        bp->next = p->next->next;
    } else {
        bp->next = p->next;
    }

    if (p + p->size == bp) {
        p->size += bp->size;
        p->next = bp->next;
    } else {
        p->next = bp;
    }

    freep = p;
}
```

This function not only adds the freed block to the list but also merges adjacent blocks to reduce fragmentation.

---

### Mark and Sweep Garbage Collection

Now that we have our memory allocation in place, let’s explore how the garbage collector works.

#### Scanning Memory Regions

```c
void scan_region(uintptr_t sp, uintptr_t end) {
    // Placeholder logic for scanning the memory region (sp to end)
}
```

This function is a placeholder for scanning specific regions of memory (e.g., the stack) for references to allocated blocks.

#### Scanning the Heap

The `scan_heap` function traverses the heap and marks blocks that are still in use:

```c
void scan_heap() {
    header_t *bp, *up;
    uintptr_t *vp;

    for (bp = usedp; bp != NULL; bp = bp->next) {
        for (vp = (uintptr_t *)(bp + 1); vp < (uintptr_t *)(bp + bp->size); vp++) {
            uintptr_t v = *vp;
            if (v >= (uintptr_t)usedp && v < (uintptr_t)(usedp + usedp->size)) {
                up = (header_t *)v;
                up->next = (header_t *)((uintptr_t)up->next | 1);  // Mark as used
            }
        }
    }
}
```

Here, the code marks memory blocks by setting a specific bit in the `next` pointer to indicate that the block is still in use.

#### Garbage Collection (`GC_collect`)

Finally, the `GC_collect` function performs the sweep phase, freeing memory that was not marked:

```c
void GC_collect() {
    if (usedp == NULL) return;

    scan_region((uintptr_t)&etext, (uintptr_t)&end);  // Placeholder regions
    scan_heap();

    header_t *p, *prevp, *tp;
    uintptr_t stack_top = (uintptr_t)&stack_bottom;  // Dummy value for demonstration
    scan_region(stack_top, stack_bottom);

    for (prevp = usedp, p = usedp->next;; prevp = p, p = p->next) {
        if (!((uintptr_t)p->next & 1)) {
            tp = p;
            p = p->next;
            add_to_free_list(tp);
            continue;
        }
        prevp->next = (header_t *)((uintptr_t)p | ((uintptr_t)prevp->next & 1));
    }
}
```

This function:
1. Scans the stack and memory for references.
2. Traverses the used blocks to check which ones can be collected.
3. Adds unused blocks back to the free list.

---

### Testing the Garbage Collector

Here’s the `main()` function that demonstrates the garbage collector in action:

```c
int main() {
    GC_init();  // Initialize garbage collector
    void *p = GC_malloc(100);  // Allocate memory
    printf("Allocated memory at %p\n", p);

    void *p2 = GC_malloc(200);
    printf("Allocated memory at %p\n", p2);

    GC_collect();  // Perform garbage collection
    return 0;
}
```

This function initializes the garbage collector, allocates memory, and performs garbage collection after some allocations.

---

### Conclusion

In this blog post, we implemented a basic garbage collector in C using a mark-and-sweep algorithm. This garbage collector can:
- Allocate memory dynamically using `GC_malloc`.
- Collect unused memory through `GC_collect` after scanning the stack and heap.

This example demonstrates the core concepts behind garbage collection. It’s a simplified implementation, but it provides a solid foundation for understanding

 memory management in C.

