#include "anal.h"

int GetVTableIndexFromThunk(void *pMethodRealPtr) noexcept {
    if (!pMethodRealPtr) return -1;

    auto *code = reinterpret_cast<const uint8_t *>(pMethodRealPtr);

    // mov rax, qword ptr [rcx] -> 48 8B 01
    if (code[0] != 0x48 || code[1] != 0x8B || code[2] != 0x01) {
        return -1;
    }

    // JMP analysis.

    // Option A: jmp qword ptr [rax] -> FF 20
    if (code[3] == 0xFF && code[4] == 0x20) {
        return 0;
    }

    // Option B: jmp qword ptr [rax + disp8] -> FF 60 XX
    // Is used for interface's first 15 methods.
    if (code[3] == 0xFF && code[4] == 0x60) {
        int8_t disp8 = static_cast<int8_t>(code[5]);
        if (disp8 % 8 == 0) {
            return disp8 / 8;
        }
    }

    // Option C: jmp qword ptr [rax + disp32] -> FF A0 XX XX XX XX
    // For large interfaces.
    if (code[3] == 0xFF && code[4] == 0xA0) {
        int32_t disp32 = *reinterpret_cast<const int32_t *>(&code[5]);
        if (disp32 % 8 == 0) {
            return disp32 / 8;
        }
    }

    return -1; // Unknown JMP behaviour.
}

void *FollowJmpThunk(void *pMethod) {
    if (!pMethod) return nullptr;

    unsigned char *code = reinterpret_cast<unsigned char *>(pMethod);

    // 0xE9 - instruction code for JMP rel32 in x86/x64
    if (code[0] == 0xE9) {
        int32_t offset = *reinterpret_cast<int32_t *>(code + 1);
        void *targetAddress = code + 5 + offset;

        return targetAddress;
    }

    // If it is not JMP thunk, return the original address.
    return pMethod;
}

void printIndex(const char *interfaceName, int index) {
    if (index != -1) {
        std::cout << interfaceName << ".method[" << index << "]" << std::endl;
    } else {
        std::cout << "ERROR" << std::endl;
    }
}
