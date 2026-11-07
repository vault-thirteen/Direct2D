# Microsoft Direct2D API in Go language

This is a simple test of the `Direct2D` API made in `Go` programming language.

This project shows the level of difficulty of using in the `Go` language a 
closed-source API made by the _Microsoft_ corporation written in an old `C++` 
language. 

## Difficulties

A lot of aspects of the `C++` language physically can not be ported into the 
`Go` language, such as numerous macros, unions, calling convention, virtual 
method tables (VMT) and many other things. 

Moreover, _Microsoft_'s architecture is very old or, better to say, ancient. 
So ancient that window creation and message reading must be performed from the 
same O.S. thread. 

Moreover, you are not allowed to send anything to standard output streams from 
the same thread where message reading is done, otherwise your application will 
simply crash. 

Let alone the difficulty of writing wrapper functions for calling functions of 
DLL files. The standard `mkwinsyscall` utility provided by `Go` language is 
useless for writing class methods and interface methods.

If developers of `Go` language want to make their language popular, they should 
at least provide propper tools for writing wrappers around `C++` class methods. 
Otherwise, this language will be lying in dusty corners of underground 
warehouses of history.

Another annoying fact is that _Microsoft_ corporation has deliberately cut out 
function names from their debug library files, such as `d2d1.pdb` for the 
`d2d1.dll`. They do not want anyone to touch their "precious" APIs which 
literally were made with technologies of the previous century.

## Calling Convention

C++ language on Windows operating systems uses a non-standard calling 
convention for class methods, i.e. the so-called "member functions".

When porting C++ code into Go language, class method functions which return a 
struct must use a special trick to catch the result. This rule applies even 
when the struct's size does not exceed the size of a CPU register ! An example 
is following.

```cpp
D2D1_SIZE_F ID2D1Bitmap::GetSize();
D2D1_SIZE_U ID2D1Bitmap::GetPixelSize();
```

The above quoted C++ class methods `GetSize` and `GetPixelSize` return a small 
struct having the size of 8 bytes (i.e. 64 bits), each of which stores two 
4-byte (i.e. 32-bit) values. Despite the fact that whole struct can be placed 
in a CPU register, C++ passes them by pointer. To catch them in Go language, 
use the following trick. 

```go
func (b *ID2D1Bitmap) GetSize() (sf D2D1_SIZE_F) {
    _, _, _ = syscall.SyscallN(b.Vtbl.GetSize, uintptr(unsafe.Pointer(b)), uintptr(unsafe.Pointer(&sf)))
return sf
}

func (b *ID2D1Bitmap) GetPixelSize() (su D2D1_SIZE_U) {
    _, _, _ = syscall.SyscallN(b.Vtbl.GetPixelSize, uintptr(unsafe.Pointer(b)), uintptr(unsafe.Pointer(&su)))
return su
}
```

Note that the return value is not read from `r1` and `r2` outputs of the 
`syscall.SyscallN` function. The `r1` output represents the `RAX` CPU register,
and the `r2` output value represents the `RDX` CPU register. Both registers 
are ignored here.

```go
func SyscallN(p uintptr, args ...uintptr) (r1, r2 uintptr, err Errno) {
	return syscalln(p, uintptr(len(args)), args...)
}
```

To catch  the returned value, use the additional input argument, which is 
passed by pointer, not by value !

Do note that all the above-mentioned facts are true for modern CPUs having 
64-bit instruction sets, the so-called Intel x86-64 and AMD64 instruction set 
architecture (ISA). 

## Floating Point Values

To send floating point values with `SyscallN` function the simple wrapping 
into an `uintptr` is not enough. To pass the `float32` type, wrap it with 
`math.Float32bits`. Example is following. 

```go
func (wrt *ID2D1HwndRenderTarget) DrawBitmap(bitmap *ID2D1Bitmap, destinationRectangle *D2D1_RECT_F, opacity FLOAT, interpolationMode D2D1_BITMAP_INTERPOLATION_MODE, sourceRectangle *D2D1_RECT_F) {
    _, _, _ = syscall.SyscallN(wrt.Vtbl.DrawBitmap, uintptr(unsafe.Pointer(wrt)), uintptr(unsafe.Pointer(bitmap)), uintptr(unsafe.Pointer(destinationRectangle)), uintptr(math.Float32bits(opacity)), uintptr(interpolationMode), uintptr(unsafe.Pointer(sourceRectangle)))
}
```

To receive a floating point value from `SyscallN` function, use the 
`math.Float32frombits` helper function. 

Pay attention to the fact that floating point values are returned into the 
second `r2` output of `SyscallN` function. This means that in order to get 
`float32` values, you should use something similar to this code:

```go
math.Float32frombits((uint64(r2))
```

If you use `float64` type, then change `Float32bits` to `Float64bits` and 
`Float32frombits` to `Float64frombits`.

## Longevity

All the above-mentioned facts may change dramatically soon, when the Intel® 
Advanced Performance Extensions (Intel® APX) come out. The APX instruction 
set may change the process of working with CPU registers so much that all the 
previous knowledge may become obsolete. 

## Status

This project is an experiment. The future of it is unknown.
