#include <windows.h>
#include <iostream>
#include <sstream>
#include <string>
#include <d2d1.h>
#include <d2d1_2.h>
#include <wincodec.h>
#include <vector>
#include <fstream>
#include <iostream>

#pragma comment(lib, "d2d1.lib")
#pragma comment(lib, "windowscodecs.lib")

class DemoApp {
    HWND m_hwnd;
    ID2D1Factory *m_pDirect2DFactory;
    ID2D1HwndRenderTarget *m_pRenderTarget;
    ID2D1Bitmap *m_pBitmap;

public:
    DemoApp() : m_hwnd(nullptr), m_pDirect2DFactory(nullptr), m_pRenderTarget(nullptr), m_pBitmap(nullptr) {
    }

    ~DemoApp() {
        SafeRelease(&m_pDirect2DFactory);
        SafeRelease(&m_pRenderTarget);
        SafeRelease(&m_pBitmap);
    }

    HRESULT Initialize(HINSTANCE hInstance);

    HRESULT RunMessageLoop();

    static void LogError(const std::string &text, HRESULT hr);

private:
    HRESULT CreateDeviceIndependentResources();

    HRESULT CreateDeviceResources();

    void DiscardDeviceResources();

    HRESULT LoadBitmapFromFile(PCWSTR filePath, int width, int height);

    HRESULT Render();

    HRESULT OnResize(UINT width, UINT height);

    static LRESULT CALLBACK WndProc(HWND hWnd, UINT message, WPARAM wParam, LPARAM lParam);

    template<class T>
    inline void SafeRelease(T **ppT) {
        if (*ppT) {
            (*ppT)->Release();
            *ppT = nullptr;
        }
    }
};

HRESULT DemoApp::Initialize(HINSTANCE hInstance) {
    HRESULT hr = CoInitialize(nullptr);
    if (FAILED(hr)) return hr;

    hr = CreateDeviceIndependentResources();
    if (FAILED(hr)) return hr;

    WNDCLASSEX wcex = {sizeof(WNDCLASSEX)};
    wcex.style = CS_HREDRAW | CS_VREDRAW;
    wcex.lpfnWndProc = DemoApp::WndProc;
    wcex.cbClsExtra = 0;
    wcex.cbWndExtra = sizeof(LONG_PTR);
    wcex.hInstance = hInstance;
    wcex.hCursor = LoadCursor(nullptr, IDC_ARROW);
    wcex.lpszClassName = L"Direct2DBitmapApp";
    RegisterClassEx(&wcex);

    m_hwnd = CreateWindow(
        L"Direct2DBitmapApp", L"Direct2D Bitmap Render",
        WS_OVERLAPPEDWINDOW, CW_USEDEFAULT, CW_USEDEFAULT,
        800, 600, nullptr, nullptr, hInstance, this
    );
    if (!m_hwnd) {
        return HRESULT_FROM_WIN32(GetLastError());
    }

    ShowWindow(m_hwnd, SW_SHOWNORMAL);

    BOOL bRet = UpdateWindow(m_hwnd);
    if (!bRet) {
        return HRESULT_FROM_WIN32(GetLastError());
    }

    return S_OK;
}

HRESULT DemoApp::CreateDeviceIndependentResources() {
    D2D1_FACTORY_OPTIONS factoryOptions;
    factoryOptions.debugLevel = D2D1_DEBUG_LEVEL_INFORMATION;
    REFIID riid = __uuidof(ID2D1Factory2);

    return D2D1CreateFactory(D2D1_FACTORY_TYPE_SINGLE_THREADED, riid, &factoryOptions,
                             reinterpret_cast<void **>(&m_pDirect2DFactory));
}

HRESULT DemoApp::RunMessageLoop() {
    MSG msg;
    BOOL bRet = GetMessage(&msg, nullptr, 0, 0);
    while (bRet != 0) {
        if (bRet < 0) {
            return HRESULT_FROM_WIN32(GetLastError());
        }

        TranslateMessage(&msg);
        DispatchMessage(&msg);
        bRet = GetMessage(&msg, nullptr, 0, 0);
    }

    return S_OK;
}

std::string PCWSTRToString(PCWSTR pcwstr) {
    if (!pcwstr) return "";

    // Determine the required size
    size_t sizeNeeded = 0;
    wcstombs_s(&sizeNeeded, nullptr, 0, pcwstr, 0);

    // Allocate the string buffer
    std::string result(sizeNeeded - 1, '\0');

    // Do the conversion
    wcstombs_s(nullptr, &result[0], sizeNeeded, pcwstr, sizeNeeded - 1);

    return result;
}

std::vector<char> readFileToByteArray(const std::string &filename) {
    std::ifstream file(filename, std::ios::binary | std::ios::ate);

    if (!file.is_open()) {
        throw std::runtime_error("Failed to open file: " + filename);
    }

    std::streamsize size = file.tellg();

    file.seekg(0, std::ios::beg);

    std::vector<char> buffer(size);

    if (!file.read(buffer.data(), size)) {
        throw std::runtime_error("Failed to read file completely: " + filename);
    }

    return buffer;
}

HRESULT DemoApp::LoadBitmapFromFile(PCWSTR filePath, int width, int height) {
    std::vector<char> bytes;
    char *srcData = nullptr;
    try {
        bytes = readFileToByteArray(PCWSTRToString(filePath));
        srcData = bytes.data();
        std::cout << "Successfully read " << bytes.size() << " bytes.\n";
    } catch (const std::exception &e) {
        std::cerr << "Error: " << e.what() << '\n';
        return -1;
    }

    auto bitmapProperties = D2D1_BITMAP_PROPERTIES{
        .pixelFormat = D2D1_PIXEL_FORMAT{
            .format = DXGI_FORMAT_B8G8R8A8_UNORM,
            .alphaMode = D2D1_ALPHA_MODE_PREMULTIPLIED,
        },
        .dpiX = 96.0f,
        .dpiY = 96.0f,
    };

    HRESULT hr = m_pRenderTarget->CreateBitmap(
        D2D1_SIZE_U{.width = UINT32(width), .height = UINT32(height)},
        srcData,
        width * 4,
        &bitmapProperties,
        &m_pBitmap
    );
    if (FAILED(hr)) {
        return hr;
    }

    return S_OK;
}

HRESULT DemoApp::Render() {
    HRESULT hr = CreateDeviceResources();
    if (FAILED(hr)) {
        LogError("CreateDeviceResources", hr);
        return hr;
    }

    auto state = m_pRenderTarget->CheckWindowState();
    if (state != D2D1_WINDOW_STATE_NONE) {
        return S_OK;
    }

    m_pRenderTarget->BeginDraw();
    m_pRenderTarget->SetTransform(D2D1::Matrix3x2F::Identity());
    m_pRenderTarget->Clear(D2D1::ColorF(D2D1::ColorF::Gray));

    if (m_pBitmap) {
        D2D1_SIZE_F size = m_pBitmap->GetSize();
        D2D1_RECT_F destinationRect = D2D1::RectF(0.0f, 0.0f, size.width, size.height);
        m_pRenderTarget->DrawBitmap(m_pBitmap, destinationRect);
    }

    hr = m_pRenderTarget->EndDraw();
    if (FAILED(hr)) {
        LogError("EndDraw", hr);
        if (hr == D2DERR_RECREATE_TARGET) {
            DiscardDeviceResources();
        } else {
            return hr;
        }
    }

    return S_OK;
}

HRESULT DemoApp::CreateDeviceResources() {
    HRESULT hr = S_OK;
    if (!m_pRenderTarget) {
        RECT rc;
        if (!GetClientRect(m_hwnd, &rc)) {
            return HRESULT_FROM_WIN32(GetLastError());
        }

        D2D1_SIZE_U size = D2D1::SizeU(rc.right - rc.left, rc.bottom - rc.top);


        D2D1_RENDER_TARGET_PROPERTIES rtp = D2D1::RenderTargetProperties();
        D2D1_PIXEL_FORMAT pixelFormat = {DXGI_FORMAT_R8G8B8A8_UNORM, D2D1_ALPHA_MODE_PREMULTIPLIED};
        rtp.pixelFormat = pixelFormat;

        D2D1_HWND_RENDER_TARGET_PROPERTIES wrtp = D2D1::HwndRenderTargetProperties(
            m_hwnd, size, D2D1_PRESENT_OPTIONS_NONE);

        hr = m_pDirect2DFactory->CreateHwndRenderTarget(&rtp, &wrtp, &m_pRenderTarget);
        if (FAILED(hr)) {
            return hr;
        }

        hr = LoadBitmapFromFile(L"pixels.bin", 900, 506);
        if (FAILED(hr)) {
            return hr;
        }
    }

    return S_OK;
}

void DemoApp::DiscardDeviceResources() {
    SafeRelease(&m_pRenderTarget);
    SafeRelease(&m_pBitmap);
}

HRESULT DemoApp::OnResize(UINT width, UINT height) {
    if (m_pRenderTarget) {
        HRESULT hr = m_pRenderTarget->Resize(D2D1::SizeU(width, height));
        if (FAILED(hr)) {
            SetLastError(HRESULT_CODE(hr));
            return hr;
        }
    }

    return S_OK;
}

LRESULT CALLBACK DemoApp::WndProc(HWND hwnd, UINT message, WPARAM wParam, LPARAM lParam) {
    // WM_CREATE message can return '-1' error.
    if (message == WM_CREATE) {
        auto pcs = reinterpret_cast<CREATESTRUCT *>(lParam);
        auto pApp = static_cast<DemoApp *>(pcs->lpCreateParams);
        SetLastError(0);
        auto res = SetWindowLongPtrW(hwnd, GWLP_USERDATA, reinterpret_cast<LONG_PTR>(pApp));
        if (res == 0) {
            DWORD error = ::GetLastError();
            if (error != ERROR_SUCCESS) {
                return -1;
            }
        }

        return 0;
    }

    auto pApp = reinterpret_cast<DemoApp *>(GetWindowLongPtrW(hwnd, GWLP_USERDATA));

    if (!pApp) {
        return DefWindowProc(hwnd, message, wParam, lParam);
    }

    HRESULT hr;
    switch (message) {
        // WM_SIZE message can not return errors.
        case WM_SIZE: {
            UINT width = LOWORD(lParam);
            UINT height = HIWORD(lParam);

            hr = pApp->OnResize(width, height);
            if (FAILED(hr)) {
                LogError("pApp->OnResize", hr);
                return -1;
            }

            return 0;
        }

        // WM_PAINT message can not return errors.
        case WM_PAINT: {
            hr = pApp->Render();
            if (FAILED(hr)) {
                LogError("pApp->Render", hr);
                return -1;
            }

            if (ValidateRect(hwnd, nullptr) == 0) {
                LogError("ValidateRect", hr);
                return -1;
            }

            return 0;
        }

        case WM_DESTROY: {
            PostQuitMessage(0);
            return 0;
        }

        default: {
            return DefWindowProc(hwnd, message, wParam, lParam);
        }
    }
}

void DemoApp::LogError(const std::string &text, HRESULT hr) {
    std::stringstream buf;
    buf << "Error " << hr << " : " << GetLastError() << " : " << text << std::endl;
    std::cout << buf.str();
}

int WINAPI WinMain(HINSTANCE hInstance, HINSTANCE, LPSTR, int) {
    AllocConsole();
    freopen_s(reinterpret_cast<FILE **>(stdout), "CONOUT$", "w", stdout);

    HeapSetInformation(nullptr, HeapEnableTerminationOnCorruption, nullptr, 0);

    HRESULT hr;
    hr = CoInitialize(nullptr);
    //hr = CoInitializeEx(nullptr, COINIT_MULTITHREADED);
    if (FAILED(hr)) {
        return hr;
    }

    DemoApp app;
    hr = app.Initialize(hInstance);
    if (FAILED(hr)) {
        return hr;
    }


    hr = app.RunMessageLoop();
    if (FAILED(hr)) {
        return hr;
    }

    CoUninitialize();

    return 0;
}
