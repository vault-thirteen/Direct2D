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

## Status

This project is an experiment. The future of it is unknown.
