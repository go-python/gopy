// Argument conversion shared by the pybind11 and nanobind backends, spliced
// into each one's generated .cpp in place of @ARG_HELPERS@ (see
// cxxArgHelpers in pybind11.go).  Each binding takes its arguments as raw
// python objects and converts them here, the same way pybindgen's generated
// code does (PyArg_ParseTuple, with the same format code per C type, plus
// pybindgen's own range checks), so that a wrongly-typed argument raises the
// same exception, with the same message, under every backend -- rather than
// pybind11's or nanobind's own "incompatible function arguments" TypeError.
// argn is the argument's 1-based position, as PyArg_ParseTuple counts it.
//
// gopy_raise must be defined first: it throws the C++ exception that hands
// the python error already set back to python (pybind11 and nanobind each
// spell that differently).  Note that this text is spliced into a python
// string literal in build.py, so it must not contain any backslashes.

// the type name PyArg_ParseTuple reports for o
static const char* gopy_arg_type(PyObject* o) {
    return o == Py_None ? "None" : Py_TYPE(o)->tp_name;
}

static void gopy_arg_error(PyObject* exc, const char* msg) {
    PyErr_SetString(exc, msg);
    gopy_raise();
}

// PyArg_ParseTuple's "s"
static const char* gopy_arg_str(PyObject* o, int argn) {
    if (!PyUnicode_Check(o)) {
        PyErr_Format(PyExc_TypeError, "argument %d must be str, not %.50s", argn, gopy_arg_type(o));
        gopy_raise();
    }
    Py_ssize_t n;
    const char* s = PyUnicode_AsUTF8AndSize(o, &n);
    if (!s) {
        gopy_raise();
    }
    if ((Py_ssize_t)strlen(s) != n) {
        gopy_arg_error(PyExc_ValueError, "embedded null character");
    }
    return s;  // borrowed from o, which outlives the call it is passed to
}

// PyArg_ParseTuple's "L"
static long long gopy_arg_L(PyObject* o) {
    long long v = PyLong_AsLongLong(o);
    if (v == -1 && PyErr_Occurred()) {
        gopy_raise();
    }
    return v;
}

// PyArg_ParseTuple's "K"
static unsigned long long gopy_arg_K(PyObject* o, int argn) {
    if (!PyLong_Check(o)) {
        PyErr_Format(PyExc_TypeError, "argument %d must be int, not %.50s", argn, gopy_arg_type(o));
        gopy_raise();
    }
    return PyLong_AsUnsignedLongLongMask(o);
}

// PyArg_ParseTuple's "i"
static int gopy_arg_i(PyObject* o) {
    long v = PyLong_AsLong(o);
    if (v == -1 && PyErr_Occurred()) {
        gopy_raise();
    }
    if (v > INT_MAX) {
        gopy_arg_error(PyExc_OverflowError, "signed integer is greater than maximum");
    }
    if (v < INT_MIN) {
        gopy_arg_error(PyExc_OverflowError, "signed integer is less than minimum");
    }
    return (int)v;
}

// PyArg_ParseTuple's "i", then pybindgen's own check, for the C types it
// parses that way despite being narrower than int (it checks only the upper
// bound).
static int gopy_arg_i_max(PyObject* o, int max) {
    int v = gopy_arg_i(o);
    if (v > max) {
        gopy_arg_error(PyExc_ValueError, "Out of range");
    }
    return v;
}

// PyArg_ParseTuple's "I"
static unsigned int gopy_arg_I(PyObject* o) {
    unsigned long v = PyLong_AsUnsignedLongMask(o);
    if (v == (unsigned long)-1 && PyErr_Occurred()) {
        gopy_raise();
    }
    return (unsigned int)v;
}

// PyArg_ParseTuple's "d" (and, narrowed, "f")
static double gopy_arg_d(PyObject* o) {
    double v = PyFloat_AsDouble(o);
    if (v == -1.0 && PyErr_Occurred()) {
        gopy_raise();
    }
    return v;
}

// pybindgen's bool: PyArg_ParseTuple's "O", then PyObject_IsTrue
static char gopy_arg_bool(PyObject* o) {
    int v = PyObject_IsTrue(o);
    if (v < 0) {
        gopy_raise();
    }
    return (char)v;
}

// as the cffi backend checks a callback argument (pybindgen can't take one)
static void gopy_arg_callable(PyObject* o, int argn) {
    if (!PyCallable_Check(o)) {
        PyErr_Format(PyExc_TypeError, "argument %d must be callable, not %.50s", argn, Py_TYPE(o)->tp_name);
        gopy_raise();
    }
}
