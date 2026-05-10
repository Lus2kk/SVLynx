cmake_minimum_required(VERSION 3.16)
project(voice_server)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

include_directories(.)

add_executable(voice_server voice_server.cpp)

target_link_libraries(voice_server pthread)

if(APPLE)
    target_link_libraries(voice_server "-framework CoreFoundation")
endif()