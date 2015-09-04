// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#ifndef CGOPY_SEQ_CPY_H
#define CGOPY_SEQ_CPY_H 1

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#if __GNUC__ >= 4
#  define CGOPY_HASCLASSVISIBILITY
#endif

#if defined(CGOPY_HASCLASSVISIBILITY)
#  define CGOPY_IMPORT __attribute__((visibility("default")))
#  define CGOPY_EXPORT __attribute__((visibility("default")))
#  define CGOPY_LOCAL  __attribute__((visibility("hidden")))
#else
#  define CGOPY_IMPORT
#  define CGOPY_EXPORT
#  define CGOPY_LOCAL
#endif

#define CGOPY_API CGOPY_EXPORT

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
	uint8_t *Data;
	int64_t  Len;
} cgopy_seq_bytearray;

CGOPY_API
cgopy_seq_bytearray
cgopy_seq_bytearray_new(int64_t len);

CGOPY_API
void
cgopy_seq_bytearray_free(cgopy_seq_bytearray arr);

typedef struct {
	uint8_t *buf;
	uint32_t off;
	uint32_t len;
	uint32_t cap;

	// TODO(hyangah): have it as a separate field outside mem?
	//pinned* pinned;
} *cgopy_seq_buffer;

CGOPY_API
cgopy_seq_buffer
cgopy_seq_buffer_new(void);

CGOPY_API
void
cgopy_seq_buffer_free(cgopy_seq_buffer buf);

CGOPY_API
uint8_t*
cgopy_seq_buffer_data(cgopy_seq_buffer buf);

CGOPY_API
size_t
cgopy_seq_buffer_len(cgopy_seq_buffer buf);

CGOPY_API
int8_t
cgopy_seq_buffer_read_bool(cgopy_seq_buffer buf);

CGOPY_API
int8_t
cgopy_seq_buffer_read_int8(cgopy_seq_buffer buf);

CGOPY_API
int16_t
cgopy_seq_buffer_read_int16(cgopy_seq_buffer buf);

CGOPY_API
int32_t
cgopy_seq_buffer_read_int32(cgopy_seq_buffer buf);

CGOPY_API
int64_t
cgopy_seq_buffer_read_int64(cgopy_seq_buffer buf);

CGOPY_API
float
cgopy_seq_buffer_read_float32(cgopy_seq_buffer buf);

CGOPY_API
double
cgopy_seq_buffer_read_float64(cgopy_seq_buffer buf);

CGOPY_API
cgopy_seq_bytearray
cgopy_seq_buffer_read_bytearray(cgopy_seq_buffer buf);

CGOPY_API
void
cgopy_seq_buffer_write_bool(cgopy_seq_buffer buf, int8_t v);

CGOPY_API
void
cgopy_seq_buffer_write_int8(cgopy_seq_buffer buf, int8_t v);

CGOPY_API
void
cgopy_seq_buffer_write_int16(cgopy_seq_buffer buf, int16_t v);

CGOPY_API
void
cgopy_seq_buffer_write_int32(cgopy_seq_buffer buf, int32_t v);

CGOPY_API
void
cgopy_seq_buffer_write_int64(cgopy_seq_buffer buf, int64_t v);

CGOPY_API
void
cgopy_seq_buffer_write_float32(cgopy_seq_buffer buf, float v);

CGOPY_API
void
cgopy_seq_buffer_write_float64(cgopy_seq_buffer buf, double v);

CGOPY_API
void
cgopy_seq_buffer_write_bytearray(cgopy_seq_buffer buf, cgopy_seq_bytearray v);



#endif /* !CGOPY_SEQ_CPY_H */
