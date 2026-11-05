cmake_minimum_required(VERSION 4.4)
project(Project)

set(CMAKE_CXX_STANDARD 26)

add_executable(Project anal.h anal.cpp main.cpp)
target_compile_definitions(Project PRIVATE UNICODE _UNICODE)

if (MSVC)
    target_link_options(Project PRIVATE "/SUBSYSTEM:WINDOWS")
endif ()
