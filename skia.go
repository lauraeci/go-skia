/*
 * Copyright 2019 Tapir Liu.
 *
 * Copyright 2015 Google Inc.
 *
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */
package skia

/*
//#cgo LDFLAGS: -L${SRCDIR}/skia/lib/shared
//#cgo LDFLAGS: -Wl,-rpath,${SRCDIR}/skia/lib/shared
//#cgo LDFLAGS: -lskia
#cgo LDFLAGS: -L${SRCDIR}/skia/lib/static
#cgo LDFLAGS: -Wl,-rpath,${SRCDIR}/skia/lib/static -lskia
#cgo LDFLAGS: -lskia -lstdc++ -lwebp -lwebpmux -lpng -lpng16 -ljpeg -lz -lm -ldl -luuid -lfreetype -lfontconfig -lexpat
#cgo CFLAGS: -I./skia/include/c
#include "sk_canvas.h"
#include "sk_data.h"
#include "sk_image.h"
#include "sk_paint.h"
#include "sk_path.h"
#include "sk_surface.h"
*/
import "C"

import (
	"fmt"
	"io"
	"math"
	"runtime"
	"unsafe"
)

const (
	radian2degrees = 180.0 / math.Pi
	degree2radians = math.Pi / 180.0
)

// TODO(stephana): Add proper documentation to the types defined here.

//////////////////////////////////////////////////////////////////////////
// Surface
//////////////////////////////////////////////////////////////////////////
type Surface struct {
	ptr *C.sk_surface_t
}

// func NewRasterSurface(width, height int32, alphaType AlphaType) (*Surface, error) {
func NewRasterSurface(imgInfo *ImageInfo) (*Surface, error) {
	ptr := C.sk_surface_new_raster(imgInfo.cPointer(), (*C.sk_surfaceprops_t)(nil))
	if ptr == nil {
		return nil, fmt.Errorf("Unable to create raster surface.")
	}

	ret := &Surface{ptr: ptr}
	runtime.SetFinalizer(ret, func(s *Surface) {
		C.sk_surface_unref(s.ptr)
	})
	return ret, nil
}

func NewDirectRasterSurface(imgInfo *ImageInfo, pixels uintptr, rowBytes int) (*Surface, error) {
	ptr := C.sk_surface_new_raster_direct(imgInfo.cPointer(), unsafe.Pointer(pixels), C.size_t(rowBytes), (*C.sk_surfaceprops_t)(nil))
	if ptr == nil {
		return nil, fmt.Errorf("Unable to create direct raster surface.")
	}

	ret := &Surface{ptr: ptr}
	runtime.SetFinalizer(ret, func(s *Surface) {
		C.sk_surface_unref(s.ptr)
	})
	return ret, nil
}

func (s *Surface) Canvas() *Canvas {
	return &Canvas{
		ptr:             C.sk_surface_get_canvas(s.ptr),
		keepParentAlive: s,
	}
}

func (s *Surface) Image() *Image {
	ret := &Image{
		ptr:             C.sk_surface_new_image_snapshot(s.ptr),
		keepParentAlive: s,
	}
	runtime.SetFinalizer(ret, func(i *Image) {
		C.sk_image_unref(i.ptr)
	})
	return ret
}

//////////////////////////////////////////////////////////////////////////
// Image
//////////////////////////////////////////////////////////////////////////
type Image struct {
	ptr             *C.sk_image_t
	keepParentAlive *Surface
}

func (i *Image) WritePNG(w io.Writer) error {
	data := C.sk_image_encode(i.ptr)
	defer C.sk_data_unref(data)

	dataPtr := C.sk_data_get_data(data)
	dataSize := C.sk_data_get_size(data)
	byteSlice := C.GoBytes(dataPtr, C.int(dataSize))
	_, err := w.Write(byteSlice)
	if err != nil {
		return err
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////
// Canvas
//////////////////////////////////////////////////////////////////////////
type Canvas struct {
	ptr             *C.sk_canvas_t
	keepParentAlive *Surface
}

func (c *Canvas) Save() {
	C.sk_canvas_save(c.ptr)
}

func (c *Canvas) Restore() {
	C.sk_canvas_restore(c.ptr)
}

func (c *Canvas) Translate(dx, dy float64) {
	C.sk_canvas_translate(c.ptr, C.float(dx), C.float(dy))
}

func (c *Canvas) Scale(sx, sy float64) {
	C.sk_canvas_scale(c.ptr, C.float(sx), C.float(sy))
}

func (c *Canvas) Rotate(radians float64) {
	C.sk_canvas_rotate_radians(c.ptr, C.float(radians))
}

func (c *Canvas) DrawPaint(paint *Paint) {
	C.sk_canvas_draw_paint(c.ptr, paint.ptr)
}

func (c *Canvas) DrawCircle(x, y, r float32, paint *Paint) {
	C.sk_canvas_draw_circle(c.ptr, C.float(x), C.float(y), C.float(r), paint.ptr)
}

func (c *Canvas) DrawOval(rect *Rect, paint *Paint) {
	// C.sk_canvas_draw_oval(c.ptr, (*C.sk_rect_t)(unsafe.Pointer(rect)), (*C.sk_paint_t)(paint.ptr))
	C.sk_canvas_draw_oval(c.ptr, rect.cPointer(), paint.ptr)
}

func (c *Canvas) DrawRect(rect *Rect, paint *Paint) {
	// C.sk_canvas_draw_rect(c.ptr, (*C.sk_rect_t)(unsafe.Pointer(rect)), (*C.sk_paint_t)(paint.ptr))
	C.sk_canvas_draw_rect(c.ptr, rect.cPointer(), paint.ptr)
}

func (c *Canvas) DrawPath(path *Path, paint *Paint) {
	// C.sk_canvas_draw_path(c.ptr, (*C.sk_path_t)(path.ptr), (*C.sk_paint_t)(paint.ptr))
	C.sk_canvas_draw_path(c.ptr, path.ptr, paint.ptr)
}

//////////////////////////////////////////////////////////////////////////
// Paint
//////////////////////////////////////////////////////////////////////////
type Paint struct {
	ptr *C.sk_paint_t
}

func NewPaint() *Paint {
	ret := &Paint{ptr: C.sk_paint_new()}
	runtime.SetFinalizer(ret, func(p *Paint) {
		C.sk_paint_delete(p.ptr)
	})
	return ret
}

func (p *Paint) SetColor(color Color) {
	C.sk_paint_set_color(p.ptr, C.sk_color_t(color))
}

func (p *Paint) SetAntiAlias(antiAlias bool) {
	C.sk_paint_set_antialias(p.ptr, C._Bool(antiAlias))
}

func (p *Paint) SetStroke(val bool) {
	C.sk_paint_set_stroke(p.ptr, C._Bool(val))
}

func (p *Paint) SetStrokeWidth(width float32) {
	C.sk_paint_set_stroke_width(p.ptr, C.float(width))
}

//////////////////////////////////////////////////////////////////////////
// Path
//////////////////////////////////////////////////////////////////////////
type Path struct {
	ptr *C.sk_path_t
}

func NewPath() *Path {
	ret := &Path{ptr: C.sk_path_new()}
	runtime.SetFinalizer(ret, func(p *Path) {
		C.sk_path_delete(p.ptr)
	})
	return ret
}

func (p *Path) MoveTo(x, y float32) {
	C.sk_path_move_to(p.ptr, C.float(x), C.float(y))
}

func (p *Path) LineTo(x, y float32) {
	C.sk_path_line_to(p.ptr, C.float(x), C.float(y))
}

func (p *Path) ArcTo(x1, y1, x2, y2, radius float32) {
	C.sk_path_arc_to(p.ptr, C.float(x1), C.float(y1), C.float(x2), C.float(y2), C.float(radius))
}

func (p *Path) Arc(x, y, r, sAngle, eAngle float32, counterclockwise bool) {
	oval := NewRect(x-r, y-r, x+r, y+r)
	if counterclockwise {
		sAngle, eAngle = eAngle, sAngle
	}
	sweepAngle := eAngle - sAngle
	C.sk_path_add_arc(p.ptr, oval.cPointer(), C.float(sAngle)*radian2degrees, C.float(sweepAngle)*radian2degrees)
}

func (p *Path) QuadTo(x0, y0, x1, y1 float32) {
	C.sk_path_quad_to(p.ptr, C.float(x0), C.float(y0), C.float(x1), C.float(y1))
}

func (p *Path) ConicTo(x0, y0, x1, y1, w float32) {
	C.sk_path_conic_to(p.ptr, C.float(x0), C.float(y0), C.float(x1), C.float(y1), C.float(w))
}

func (p *Path) CubicTo(x0, y0, x1, y1, x2, y2 float32) {
	C.sk_path_cubic_to(p.ptr, C.float(x0), C.float(y0), C.float(x1), C.float(y1), C.float(x2), C.float(y2))
}

func (p *Path) Close() {
	C.sk_path_close(p.ptr)
}

// NewRect is a convenience function to define a Rect in a single line.
func NewRect(left, top, right, bottom float32) *Rect {
	return &Rect{
		Left:   left,
		Top:    top,
		Right:  right,
		Bottom: bottom,
	}
}

// cPointer casts the pointer to Rect to the corresponding C pointer.
func (r *Rect) cPointer() *C.sk_rect_t {
	return (*C.sk_rect_t)(unsafe.Pointer(r))
}

// cPointer casts the pointer to ImageInfo to the corresponding C pointer.
func (i *ImageInfo) cPointer() *C.sk_imageinfo_t {
	return (*C.sk_imageinfo_t)(unsafe.Pointer(i))
}

// Utility functions.
func GetDefaultColortype() ColorType {
	//return ColorType(C.sk_colortype_get_default_8888())
	return RGBA_8888_COLORTYPE
}
