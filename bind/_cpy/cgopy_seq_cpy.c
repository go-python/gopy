// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include <stdint.h>
#include <stdio.h>
#include <stdarg.h>
#include <unistd.h>
#include "_cgo_export.h"

#include "cgopy_seq_cpy.h"

void LOG_FATAL(const char* format, ...) {

	va_list arg;

	va_start (arg, format);
	vfprintf (stderr, format, arg);
	va_end   (arg);
	exit(1);
}

// mem is a simple C equivalent of seq.Buffer.
//
// Many of the allocations around mem could be avoided to improve
// function call performance, but the goal is to start simple.
typedef struct mem {
	uint8_t *buf;
	uint32_t off;
	uint32_t len;
	uint32_t cap;

	// TODO(hyangah): have it as a separate field outside mem?
	//pinned* pinned;
} mem;

// mem_ensure ensures that m has at least size bytes free.
// If m is NULL, it is created.
static mem *mem_ensure(mem *m, uint32_t size) {
	if (m == NULL) {
		m = (mem*)malloc(sizeof(mem));
		if (m == NULL) {
			LOG_FATAL("mem_ensure malloc failed");
		}
		m->cap = 0;
		m->off = 0;
		m->len = 0;
		m->buf = NULL;
		//m->pinned = NULL;
	}
	uint32_t cap = m->cap;
	if (m->cap > m->off+size) {
		return m;
	}
	if (cap == 0) {
		cap = 64;
	}
	// TODO(hyangah): consider less aggressive allocation such as
	//   cap += max(pow2round(size), 64)
	while (cap < m->off+size) {
		cap *= 2;
	}
	m->buf = (uint8_t*)realloc((void*)m->buf, cap);
	if (m->buf == NULL) {
		LOG_FATAL("mem_ensure realloc failed, off=%d, size=%d", m->off, size);
	}
	m->cap = cap;
	return m;
}

static uint32_t align(uint32_t offset, uint32_t alignment) {
	uint32_t pad = offset % alignment;
	if (pad > 0) {
		pad = alignment-pad;
	}
	return pad+offset;
}

static uint8_t *mem_read(mem *m, uint32_t size, uint32_t alignment) {
	if (size == 0) {
		return NULL;
	}
	uint32_t offset = align(m->off, alignment);

	if (m->len-offset < size) {
		LOG_FATAL("short read");
	}
	uint8_t *res = m->buf+offset;
	m->off = offset+size;
	return res;
}

uint8_t *mem_write(mem *m, uint32_t size, uint32_t alignment) {
	if (m->off != m->len) {
		LOG_FATAL("write can only append to seq, size: (off=%d, len=%d, size=%d", m->off, m->len, size);
	}
	uint32_t offset = align(m->off, alignment);
	m = mem_ensure(m, offset - m->off + size);
	uint8_t *res = m->buf+offset;
	m->off = offset+size;
	m->len = offset+size;
	return res;
}

void
mem_free(mem *m) {
	if (m==NULL) {
		return;
	}
	free((void*)m->buf);
	free((void*)m);
	m=NULL;
}

cgopy_seq_bytearray
cgopy_seq_bytearray_new(int64_t len) {
	cgopy_seq_bytearray arr;
	arr.Len = len;
	arr.Data = (uint8_t*)malloc(len);
	return arr;
}

void
cgopy_seq_bytearray_free(cgopy_seq_bytearray arr) {
	free((void*)arr.Data);
	arr.Data = NULL;
	return;
}

cgopy_seq_buffer
cgopy_seq_buffer_new(void) {
	mem *m = (mem*)malloc(sizeof(mem));
	m->buf = NULL;
	m->off = 0;
	m->len = 0;
	m->cap = 0;
	return (cgopy_seq_buffer)m;
}

void
cgopy_seq_buffer_free(cgopy_seq_buffer buf) {
	mem_free((mem*)buf);
}

#define MEM_READ(buf, ty) ((ty*)mem_read((mem*)(buf), sizeof(ty), sizeof(ty)))

int8_t
cgopy_seq_buffer_read_bool(cgopy_seq_buffer buf) {
	int8_t *v = MEM_READ(buf, int8_t);
	if (v == NULL) {
		return 0;
	}
	return *v != 0 ? 1 : 0;
}

int8_t
cgopy_seq_buffer_read_int8(cgopy_seq_buffer buf) {
	int8_t *v = MEM_READ(buf, int8_t);
	return v == NULL ? 0 : *v;
}

int16_t
cgopy_seq_buffer_read_int16(cgopy_seq_buffer buf) {
	int16_t *v = MEM_READ(buf, int16_t);
	return v == NULL ? 0 : *v;
}

int32_t
cgopy_seq_buffer_read_int32(cgopy_seq_buffer buf) {
	int32_t *v = MEM_READ(buf, int32_t);
	return v == NULL ? 0 : *v;
}

int64_t
cgopy_seq_buffer_read_int64(cgopy_seq_buffer buf) {
	int64_t *v = MEM_READ(buf, int64_t);
	return v == NULL ? 0 : *v;
}

float
cgopy_seq_buffer_read_float32(cgopy_seq_buffer buf) {
	float *v = MEM_READ(buf, float);
	return v == NULL ? 0 : *v;
}

double
cgopy_seq_buffer_read_float64(cgopy_seq_buffer buf) {
	double *v = MEM_READ(buf, double);
	return v == NULL ? 0 : *v;
}

cgopy_seq_bytearray
cgopy_seq_buffer_read_bytearray(cgopy_seq_buffer buf) {
	cgopy_seq_bytearray arr;
	arr.Data = NULL;
	arr.Len = 0;
	int64_t size = cgopy_seq_buffer_read_int64(buf);
	if (size==0) {
		return arr;
	}
	arr = cgopy_seq_bytearray_new(size);
	int64_t ptr = cgopy_seq_buffer_read_int64(buf);
	arr.Data = (uint8_t*)(intptr_t)(ptr);
	return arr;
}

#define MEM_WRITE(ty) (*(ty*)mem_write((mem*)buf, sizeof(ty), sizeof(ty)))

void
cgopy_seq_buffer_write_bool(cgopy_seq_buffer buf, int8_t v) {
	MEM_WRITE(int8_t) = v ? 1 : 0;
}

void
cgopy_seq_buffer_write_int8(cgopy_seq_buffer buf, int8_t v) {
	MEM_WRITE(int8_t) = v;
}

void
cgopy_seq_buffer_write_int16(cgopy_seq_buffer buf, int16_t v) {
	MEM_WRITE(int16_t) = v;
}

void
cgopy_seq_buffer_write_int32(cgopy_seq_buffer buf, int32_t v) {
	MEM_WRITE(int32_t) = v;
}

void
cgopy_seq_buffer_write_int64(cgopy_seq_buffer buf, int64_t v) {
	MEM_WRITE(int64_t) = v;
}

void
cgopy_seq_buffer_write_float32(cgopy_seq_buffer buf, float v) {
	MEM_WRITE(float) = v;
}

void
cgopy_seq_buffer_write_float64(cgopy_seq_buffer buf, double v) {
	MEM_WRITE(double) = v;
}

void
cgopy_seq_buffer_write_bytearray(cgopy_seq_buffer buf, cgopy_seq_bytearray v) {
	// for bytearray, we pass only the (array-length, pointer) pair
	// encoded as 2 int64 values.
	// if the array length is 0, the pointer value is omitted.
	if (v.Len == 0) {
		MEM_WRITE(int64_t) = 0;
		return;
	}

	MEM_WRITE(int64_t) = v.Len;
	MEM_WRITE(int64_t) = (int64_t)(uintptr_t)v.Data;
}


