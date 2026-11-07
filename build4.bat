:: Build with normal inlining.
::go build -gcflags="all=-l=0" main.go

:: Build with aggressive inlining.
go build -gcflags="all=-l=4" main.go
