#ifndef INC_2_ANAL_H
#define INC_2_ANAL_H

#include <iostream>
#include <chrono>
#include <string>
#include <bit>
#include <cstdio>

int GetVTableIndexFromThunk(void *pMethodRealPtr) noexcept;

void *FollowJmpThunk(void *pMethod);

// For C++ 20.
template<typename MemberPtr>
void *GetMemberVoidPtr(MemberPtr ptr) noexcept {
    static_assert(sizeof(MemberPtr) == sizeof(void *),
                  "pointer-to-member is not a void*");

    return std::bit_cast<void *>(ptr);
}

void printIndex(const char *interfaceName, int index);


template<typename ReturnType, typename ClassType, typename... Args>
bool AnalyzeMethod(bool brief, void **pVTable, int idx, const char *interfaceName, const char *methodName,
                   ReturnType (ClassType::*methodPtr)(Args...)) {
    void *thunkPtr = GetMemberVoidPtr(methodPtr);
    void *realPtr = FollowJmpThunk(thunkPtr);
    auto index = GetVTableIndexFromThunk(realPtr);

    if ((index < 0) || (index != idx)) {
        std::cout << "Error in method " << interfaceName << "::" << methodName << ":" << " index=" << index << std::endl;
        std::cout << "----------------------------------------" << std::endl;
        return false;
    }

    if (brief) {
        printf("%s::[%d]%s\n", interfaceName, index, methodName);
    } else {
        std::cout << pVTable[idx] << " " << interfaceName << " (VTable)" << std::endl;
        printf("%p %s::%s (TH)\n", thunkPtr, interfaceName, methodName);
        printf("%p %s::%s (R)\n", realPtr, interfaceName, methodName);
        printIndex(interfaceName, index);
        std::cout << "----------------------------------------" << std::endl;
    }

    return true;
}

template<typename ReturnType, typename ClassType, typename... Args>
bool AnalyzeMethod_StdCall_const(bool brief, void **pVTable, int idx, const char *interfaceName, const char *methodName,
                                 ReturnType (__stdcall ClassType::*methodPtr)(Args...) const) {
    void *thunkPtr = GetMemberVoidPtr(methodPtr);
    void *realPtr = FollowJmpThunk(thunkPtr);
    auto index = GetVTableIndexFromThunk(realPtr);

    if ((index < 0) || (index != idx)) {
        std::cout << "Error in method " << interfaceName << "::" << methodName << ":" << " index=" << index << std::endl;
        std::cout << "----------------------------------------" << std::endl;
        return false;
    }

    if (brief) {
        printf("%s::[%d]%s\n", interfaceName, index, methodName);
    } else {
        std::cout << pVTable[idx] << " " << interfaceName << " (VTable)" << std::endl;
        printf("%p %s::%s (TH)\n", thunkPtr, interfaceName, methodName);
        printf("%p %s::%s (R)\n", realPtr, interfaceName, methodName);
        printIndex(interfaceName, index);
        std::cout << "----------------------------------------" << std::endl;
    }

    return true;
}


#endif //INC_2_ANAL_H
