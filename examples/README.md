# Gopiq Examples

This directory contains comprehensive examples demonstrating the features and capabilities of the gopiq image processing library.

## Files

### Sample Image
- `sample_image.png` - A colorful test image with geometric shapes (800x600 pixels)

### Example Programs

1. **`basic_example.go`** - Demonstrates basic image operations
   - Loading images from bytes
   - Grayscale conversion
   - Resizing
   - Cropping
   - Chaining operations

2. **`watermark_example.go`** - Shows text watermarking capabilities
   - Different watermark positions (top-left, top-right, bottom-left, bottom-right, center)
   - Custom styling (colors, sizes, offsets)
   - Multiple watermark styles

3. **`grayscale_example.go`** - Performance comparison of grayscale methods
   - Regular grayscale conversion
   - Fast parallel grayscale conversion
   - Custom performance options
   - Timing comparisons

4. **`resize_example.go`** - Comprehensive resize operations
   - Various size formats (thumbnail, medium, large, square, wide, tall)
   - Aspect ratio preservation
   - Chaining with other operations

5. **`crop_example.go`** - Image cropping demonstrations
   - Corner crops
   - Center crops
   - Grid-based cropping
   - Error handling for invalid crops

6. **`comprehensive_example.go`** - Complete feature demonstration
   - All major features in one example
   - Performance comparisons
   - Error handling
   - Memory efficiency
   - Different output formats

## Running the Examples

To run any example:

```bash
cd examples
go run <example_name>.go
```

For example:
```bash
go run basic_example.go
go run watermark_example.go
go run comprehensive_example.go
```

## Generated Output Files

Each example generates various output images demonstrating different operations:

### Basic Example Outputs
- `output_grayscale.png` - Grayscale version of the sample image
- `output_resized.png` - Resized to 400x300
- `output_cropped.png` - Cropped 200x150 portion

### Watermark Example Outputs
- `watermark_top_left.png` - Watermark in top-left position
- `watermark_top_right.png` - Watermark in top-right position
- `watermark_bottom_left.png` - Watermark in bottom-left position
- `watermark_bottom_right.png` - Watermark in bottom-right position
- `watermark_center.png` - Watermark in center position
- `watermark_custom_style.png` - Custom styled watermark

### Grayscale Example Outputs
- `grayscale_regular.png` - Standard grayscale conversion
- `grayscale_fast.png` - Optimized parallel grayscale conversion
- `grayscale_custom_perf.png` - Custom performance options

### Resize Example Outputs
- `resize_thumbnail.png` - 100x75 thumbnail
- `resize_medium.png` - 400x300 medium size
- `resize_large.png` - 1200x900 large size
- `resize_square.png` - 500x500 square format
- `resize_wide.png` - 800x200 wide format
- `resize_tall.png` - 200x800 tall format
- `resize_chain_grayscale.png` - Resize + grayscale chain
- `resize_aspect_ratio.png` - Aspect ratio preserved resize

### Crop Example Outputs
- `crop_top_left.png` - Top-left corner crop
- `crop_top_right.png` - Top-right corner crop
- `crop_bottom_left.png` - Bottom-left corner crop
- `crop_bottom_right.png` - Bottom-right corner crop
- `crop_center_square.png` - Center square crop
- `crop_center_rect.png` - Center rectangle crop
- `crop_small_detail.png` - Small detail crop
- `crop_wide_strip.png` - Wide strip crop
- `crop_tall_strip.png` - Tall strip crop
- `crop_chain_resize.png` - Crop + resize chain
- `crop_grid_*.png` - 4x4 grid of crops

### Comprehensive Example Outputs
- `comprehensive_basic.png` - Basic operations result
- `comprehensive_watermark_*.png` - Various watermark styles
- `comprehensive_grayscale_*.png` - Performance comparison results
- `comprehensive_pipeline.png` - Complex processing pipeline result
- `comprehensive_format.png/.jpg` - Different output formats
- `comprehensive_efficiency_*.png` - Memory efficiency test results

## Key Features Demonstrated

- **Image Loading**: Loading images from byte arrays
- **Grayscale Conversion**: Both regular and optimized parallel methods
- **Resizing**: High-quality Catmull-Rom interpolation
- **Cropping**: Precise rectangular cropping with bounds checking
- **Watermarking**: Text watermarks with customizable positioning and styling
- **Chaining**: Fluent API for combining multiple operations
- **Performance**: Parallel processing and performance optimization
- **Error Handling**: Comprehensive error checking and reporting
- **Format Support**: PNG and JPEG output formats
- **Memory Efficiency**: Object pooling and efficient memory usage

## Performance Notes

The examples include performance comparisons showing:
- Fast grayscale conversion can be 1.4x to 5x faster than regular conversion
- Parallel processing is most beneficial for larger images
- Custom performance options allow fine-tuning for specific use cases

## Error Handling

Examples demonstrate proper error handling for:
- Invalid crop coordinates
- Negative dimensions
- Empty watermark text
- Out-of-bounds operations

All examples include comprehensive error checking and informative error messages.
