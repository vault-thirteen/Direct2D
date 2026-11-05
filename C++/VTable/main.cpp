#include <windows.h>
#include <iostream>
#include <sstream>
#include <string>
#include <d2d1.h>
#include <d2d1_2.h>
#include <wincodec.h>

#include "anal.h"

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
    void AnalyseVTables();

    HRESULT CreateDeviceIndependentResources();

    HRESULT CreateDeviceResources();

    void DiscardDeviceResources();

    HRESULT LoadBitmapFromFile(ID2D1RenderTarget *pRenderTarget, IWICImagingFactory *pIWICFactory,
                               PCWSTR uri, UINT destinationWidth, UINT destinationHeight, ID2D1Bitmap **ppBitmap);

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

HRESULT DemoApp::LoadBitmapFromFile(ID2D1RenderTarget *pRenderTarget, IWICImagingFactory *pIWICFactory,
                                    PCWSTR uri, UINT destinationWidth, UINT destinationHeight, ID2D1Bitmap **ppBitmap) {
    IWICBitmapDecoder *pDecoder = nullptr;
    IWICBitmapFrameDecode *pSource = nullptr;
    IWICFormatConverter *pConverter = nullptr;

    HRESULT hr = pIWICFactory->CreateDecoderFromFilename(
        uri, nullptr, GENERIC_READ, WICDecodeMetadataCacheOnLoad, &pDecoder
    );
    if (FAILED(hr)) {
        return hr;
    }

    hr = pDecoder->GetFrame(0, &pSource);
    if (FAILED(hr)) {
        return hr;
    }

    hr = pIWICFactory->CreateFormatConverter(&pConverter);
    if (FAILED(hr)) {
        return hr;
    }

    hr = pConverter->Initialize(
        pSource, GUID_WICPixelFormat32bppPBGRA,
        WICBitmapDitherTypeNone, nullptr, 0.f, WICBitmapPaletteTypeMedianCut
    );
    if (FAILED(hr)) {
        return hr;
    }

    hr = pRenderTarget->CreateBitmapFromWicBitmap(pConverter, nullptr, ppBitmap);
    if (FAILED(hr)) {
        return hr;
    }

    SafeRelease(&pDecoder);
    SafeRelease(&pSource);
    SafeRelease(&pConverter);

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

    AnalyseVTables();

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

        // Initialising WIC to load images.
        IWICImagingFactory *pIWICFactory = nullptr;
        hr = CoCreateInstance(
            CLSID_WICImagingFactory, nullptr, CLSCTX_INPROC_SERVER,
            IID_PPV_ARGS(&pIWICFactory)
        );
        if (FAILED(hr)) {
            return hr;
        }

        hr = LoadBitmapFromFile(m_pRenderTarget, pIWICFactory, L"test.png", 0, 0, &m_pBitmap);
        if (FAILED(hr)) {
            return hr;
        }

        ULONG refCount = pIWICFactory->Release();
        if (refCount != 0) {
            return E_FAIL;
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

void DemoApp::AnalyseVTables() {
    std::cout << "Analysis of VTables" << std::endl;
    bool isBrief = true;

    // 1. Check methods of ID2D1Factory.
    printf("%p %s\n", m_pDirect2DFactory, "ID2D1Factory");
    void **ID2D1Factory_VTable = *(void ***) m_pDirect2DFactory;
    {
        // N.B. Methods which have overloaded versions, need type casting.

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 0, "ID2D1Factory", "QueryInterface",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Factory::*)(REFIID, void **)>(&ID2D1Factory::QueryInterface));

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 1, "ID2D1Factory", "AddRef", &ID2D1Factory::AddRef);

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 2, "ID2D1Factory", "Release", &ID2D1Factory::Release);

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 3, "ID2D1Factory", "ReloadSystemMetrics", &ID2D1Factory::ReloadSystemMetrics);

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 4, "ID2D1Factory", "GetDesktopDpi", &ID2D1Factory::GetDesktopDpi);

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 5, "ID2D1Factory", "CreateRectangleGeometry",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Factory::*)(const D2D1_RECT_F *, ID2D1RectangleGeometry **)>(&ID2D1Factory::CreateRectangleGeometry));

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 6, "ID2D1Factory", "CreateRoundedRectangleGeometry",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Factory::*)(const D2D1_ROUNDED_RECT *, ID2D1RoundedRectangleGeometry **)>(&ID2D1Factory::CreateRoundedRectangleGeometry));

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 7, "ID2D1Factory", "CreateEllipseGeometry",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Factory::*)(const D2D1_ELLIPSE *, ID2D1EllipseGeometry **)>(&ID2D1Factory::CreateEllipseGeometry));

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 8, "ID2D1Factory", "CreateGeometryGroup", &ID2D1Factory::CreateGeometryGroup);

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 9, "ID2D1Factory", "CreateTransformedGeometry",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Factory::*)(ID2D1Geometry *, const D2D1_MATRIX_3X2_F *, ID2D1TransformedGeometry **)>(&ID2D1Factory::CreateTransformedGeometry));

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 10, "ID2D1Factory", "CreatePathGeometry", &ID2D1Factory::CreatePathGeometry);

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 11, "ID2D1Factory", "CreateStrokeStyle",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Factory::*)(const D2D1_STROKE_STYLE_PROPERTIES *, const FLOAT *, UINT32, ID2D1StrokeStyle **)>(&ID2D1Factory::CreateStrokeStyle));

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 12, "ID2D1Factory", "CreateDrawingStateBlock",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Factory::*)(const D2D1_DRAWING_STATE_DESCRIPTION *, IDWriteRenderingParams *, ID2D1DrawingStateBlock **)>(&ID2D1Factory::CreateDrawingStateBlock));

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 13, "ID2D1Factory", "CreateWicBitmapRenderTarget",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Factory::*)(IWICBitmap *, const D2D1_RENDER_TARGET_PROPERTIES *, ID2D1RenderTarget **)>(&ID2D1Factory::CreateWicBitmapRenderTarget));

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 14, "ID2D1Factory", "CreateHwndRenderTarget",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Factory::*)(const D2D1_RENDER_TARGET_PROPERTIES *, const D2D1_HWND_RENDER_TARGET_PROPERTIES *, ID2D1HwndRenderTarget **)>(&ID2D1Factory::CreateHwndRenderTarget));

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 15, "ID2D1Factory", "CreateDxgiSurfaceRenderTarget",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Factory::*)(IDXGISurface *, const D2D1_RENDER_TARGET_PROPERTIES *, ID2D1RenderTarget **)>(&ID2D1Factory::CreateDxgiSurfaceRenderTarget));

        AnalyzeMethod(isBrief, ID2D1Factory_VTable, 16, "ID2D1Factory", "CreateDCRenderTarget", &ID2D1Factory::CreateDCRenderTarget);
    }
    std::cout << std::endl;

    // 2. Check methods of ID2D1HwndRenderTarget.
    printf("%p %s\n", m_pRenderTarget, "ID2D1HwndRenderTarget");
    void **ID2D1HwndRenderTarget_VTable = *(void ***) m_pRenderTarget;
    {
        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 0, "ID2D1HwndRenderTarget", "QueryInterface",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(REFIID, void **)>(&ID2D1HwndRenderTarget::QueryInterface));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 1, "ID2D1HwndRenderTarget", "AddRef", &ID2D1HwndRenderTarget::AddRef);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 2, "ID2D1HwndRenderTarget", "Release", &ID2D1HwndRenderTarget::Release);

        AnalyzeMethod_StdCall_const<>(isBrief, ID2D1HwndRenderTarget_VTable, 3, "ID2D1HwndRenderTarget", "GetFactory", &ID2D1HwndRenderTarget::GetFactory);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 4, "ID2D1HwndRenderTarget", "CreateBitmap",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(D2D1_SIZE_U, const void *, UINT32, const D2D1_BITMAP_PROPERTIES *, ID2D1Bitmap **)>(&ID2D1HwndRenderTarget::CreateBitmap));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 5, "ID2D1HwndRenderTarget", "CreateBitmapFromWicBitmap",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(IWICBitmapSource *, const D2D1_BITMAP_PROPERTIES *, ID2D1Bitmap **)>(&ID2D1HwndRenderTarget::CreateBitmapFromWicBitmap));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 6, "ID2D1HwndRenderTarget", "CreateSharedBitmap", &ID2D1HwndRenderTarget::CreateSharedBitmap);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 7, "ID2D1HwndRenderTarget", "CreateBitmapBrush",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(ID2D1Bitmap *, const D2D1_BITMAP_BRUSH_PROPERTIES *, const D2D1_BRUSH_PROPERTIES *, ID2D1BitmapBrush **)>(&ID2D1HwndRenderTarget::CreateBitmapBrush));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 8, "ID2D1HwndRenderTarget", "CreateSolidColorBrush",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_COLOR_F *, const D2D1_BRUSH_PROPERTIES *, ID2D1SolidColorBrush **)>(&ID2D1HwndRenderTarget::CreateSolidColorBrush));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 9, "ID2D1HwndRenderTarget", "CreateGradientStopCollection",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_GRADIENT_STOP *, UINT32, D2D1_GAMMA, D2D1_EXTEND_MODE, ID2D1GradientStopCollection **)>(&ID2D1HwndRenderTarget::CreateGradientStopCollection));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 10, "ID2D1HwndRenderTarget", "CreateLinearGradientBrush",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_LINEAR_GRADIENT_BRUSH_PROPERTIES *, const D2D1_BRUSH_PROPERTIES *, ID2D1GradientStopCollection *, ID2D1LinearGradientBrush **)>(&ID2D1HwndRenderTarget::CreateLinearGradientBrush));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 11, "ID2D1HwndRenderTarget", "CreateRadialGradientBrush",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_RADIAL_GRADIENT_BRUSH_PROPERTIES *, const D2D1_BRUSH_PROPERTIES *, ID2D1GradientStopCollection *, ID2D1RadialGradientBrush **)>(&ID2D1HwndRenderTarget::CreateRadialGradientBrush));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 12, "ID2D1HwndRenderTarget", "CreateCompatibleRenderTarget",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_SIZE_F *, const D2D1_SIZE_U *, const D2D1_PIXEL_FORMAT *, D2D1_COMPATIBLE_RENDER_TARGET_OPTIONS, ID2D1BitmapRenderTarget **)>(&ID2D1HwndRenderTarget::CreateCompatibleRenderTarget));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 13, "ID2D1HwndRenderTarget", "CreateLayer",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_SIZE_F *, ID2D1Layer **)>(&ID2D1HwndRenderTarget::CreateLayer));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 14, "ID2D1HwndRenderTarget", "CreateMesh", &ID2D1HwndRenderTarget::CreateMesh);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 15, "ID2D1HwndRenderTarget", "DrawLine", &ID2D1HwndRenderTarget::DrawLine);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 16, "ID2D1HwndRenderTarget", "DrawRectangle",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_RECT_F *, ID2D1Brush *, FLOAT, ID2D1StrokeStyle *)>(&ID2D1HwndRenderTarget::DrawRectangle));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 17, "ID2D1HwndRenderTarget", "FillRectangle",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_RECT_F *, ID2D1Brush *)>(&ID2D1HwndRenderTarget::FillRectangle));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 18, "ID2D1HwndRenderTarget", "DrawRoundedRectangle",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_ROUNDED_RECT *, ID2D1Brush *, FLOAT, ID2D1StrokeStyle *)>(&ID2D1HwndRenderTarget::DrawRoundedRectangle));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 19, "ID2D1HwndRenderTarget", "FillRoundedRectangle",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_ROUNDED_RECT *, ID2D1Brush *)>(&ID2D1HwndRenderTarget::FillRoundedRectangle));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 20, "ID2D1HwndRenderTarget", "DrawEllipse",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_ELLIPSE *, ID2D1Brush *, FLOAT, ID2D1StrokeStyle *)>(&ID2D1HwndRenderTarget::DrawEllipse));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 21, "ID2D1HwndRenderTarget", "FillEllipse",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_ELLIPSE *, ID2D1Brush *)>(&ID2D1HwndRenderTarget::FillEllipse));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 22, "ID2D1HwndRenderTarget", "DrawGeometry", &ID2D1HwndRenderTarget::DrawGeometry);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 23, "ID2D1HwndRenderTarget", "FillGeometry", &ID2D1HwndRenderTarget::FillGeometry);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 24, "ID2D1HwndRenderTarget", "FillMesh", &ID2D1HwndRenderTarget::FillMesh);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 25, "ID2D1HwndRenderTarget", "FillOpacityMask",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(ID2D1Bitmap *, ID2D1Brush *, D2D1_OPACITY_MASK_CONTENT, const D2D1_RECT_F *, const D2D1_RECT_F *)>(&ID2D1HwndRenderTarget::FillOpacityMask));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 26, "ID2D1HwndRenderTarget", "DrawBitmap",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(ID2D1Bitmap *, const D2D1_RECT_F *, FLOAT, D2D1_BITMAP_INTERPOLATION_MODE, const D2D1_RECT_F *)>(&ID2D1HwndRenderTarget::DrawBitmap));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 27, "ID2D1HwndRenderTarget", "DrawTextW",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const WCHAR *, UINT32, IDWriteTextFormat *, const D2D1_RECT_F *, ID2D1Brush *, D2D1_DRAW_TEXT_OPTIONS, DWRITE_MEASURING_MODE)>(&ID2D1HwndRenderTarget::DrawTextW));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 28, "ID2D1HwndRenderTarget", "DrawTextLayout", &ID2D1HwndRenderTarget::DrawTextLayout);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 29, "ID2D1HwndRenderTarget", "DrawGlyphRun", &ID2D1HwndRenderTarget::DrawGlyphRun);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 30, "ID2D1HwndRenderTarget", "SetTransform",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_MATRIX_3X2_F *)>(&ID2D1HwndRenderTarget::SetTransform));

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 31, "ID2D1HwndRenderTarget", "GetTransform",
                                    static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(D2D1_MATRIX_3X2_F *) const>(&ID2D1HwndRenderTarget::GetTransform));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 32, "ID2D1HwndRenderTarget", "SetAntialiasMode", &ID2D1HwndRenderTarget::SetAntialiasMode);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 33, "ID2D1HwndRenderTarget", "GetAntialiasMode",
                                    static_cast<D2D1_ANTIALIAS_MODE(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)() const>(&ID2D1HwndRenderTarget::GetAntialiasMode));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 34, "ID2D1HwndRenderTarget", "SetTextAntialiasMode", &ID2D1HwndRenderTarget::SetTextAntialiasMode);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 35, "ID2D1HwndRenderTarget", "GetTextAntialiasMode",
                                    static_cast<D2D1_TEXT_ANTIALIAS_MODE(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)() const>(&ID2D1HwndRenderTarget::GetTextAntialiasMode));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 36, "ID2D1HwndRenderTarget", "SetTextRenderingParams", &ID2D1HwndRenderTarget::SetTextRenderingParams);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 37, "ID2D1HwndRenderTarget", "GetTextRenderingParams",
                                    static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(IDWriteRenderingParams **) const>(&ID2D1HwndRenderTarget::GetTextRenderingParams));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 38, "ID2D1HwndRenderTarget", "SetTags", &ID2D1HwndRenderTarget::SetTags);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 39, "ID2D1HwndRenderTarget", "GetTags",
                                    static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(D2D1_TAG *, D2D1_TAG *) const>(&ID2D1HwndRenderTarget::GetTags));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 40, "ID2D1HwndRenderTarget", "PushLayer",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_LAYER_PARAMETERS *, ID2D1Layer *)>(&ID2D1HwndRenderTarget::PushLayer));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 41, "ID2D1HwndRenderTarget", "PopLayer", &ID2D1HwndRenderTarget::PopLayer);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 42, "ID2D1HwndRenderTarget", "Flush", &ID2D1HwndRenderTarget::Flush);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 43, "ID2D1HwndRenderTarget", "SaveDrawingState",
                                    static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(ID2D1DrawingStateBlock *) const>(&ID2D1HwndRenderTarget::SaveDrawingState));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 44, "ID2D1HwndRenderTarget", "RestoreDrawingState", &ID2D1HwndRenderTarget::RestoreDrawingState);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 45, "ID2D1HwndRenderTarget", "PushAxisAlignedClip",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_RECT_F *, D2D1_ANTIALIAS_MODE)>(&ID2D1HwndRenderTarget::PushAxisAlignedClip));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 46, "ID2D1HwndRenderTarget", "PopAxisAlignedClip", &ID2D1HwndRenderTarget::PopAxisAlignedClip);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 47, "ID2D1HwndRenderTarget", "Clear",
                      static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_COLOR_F *)>(&ID2D1HwndRenderTarget::Clear));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 48, "ID2D1HwndRenderTarget", "BeginDraw", &ID2D1HwndRenderTarget::BeginDraw);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 49, "ID2D1HwndRenderTarget", "EndDraw", &ID2D1HwndRenderTarget::EndDraw);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 50, "ID2D1HwndRenderTarget", "GetPixelFormat",
                                    static_cast<D2D1_PIXEL_FORMAT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)() const>(&ID2D1HwndRenderTarget::GetPixelFormat));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 51, "ID2D1HwndRenderTarget", "SetDpi", &ID2D1HwndRenderTarget::SetDpi);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 52, "ID2D1HwndRenderTarget", "GetDpi",
                                    static_cast<void(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(FLOAT *, FLOAT *) const>(&ID2D1HwndRenderTarget::GetDpi));

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 53, "ID2D1HwndRenderTarget", "GetSize",
                                    static_cast<D2D1_SIZE_F(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)() const>(&ID2D1HwndRenderTarget::GetSize));

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 54, "ID2D1HwndRenderTarget", "GetPixelSize",
                                    static_cast<D2D1_SIZE_U(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)() const>(&ID2D1HwndRenderTarget::GetPixelSize));

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 55, "ID2D1HwndRenderTarget", "GetMaximumBitmapSize",
                                    static_cast<UINT32(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)() const>(&ID2D1HwndRenderTarget::GetMaximumBitmapSize));

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 56, "ID2D1HwndRenderTarget", "IsSupported",
                                    static_cast<BOOL(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_RENDER_TARGET_PROPERTIES *) const>(&ID2D1HwndRenderTarget::IsSupported));

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 57, "ID2D1HwndRenderTarget", "CheckWindowState", &ID2D1HwndRenderTarget::CheckWindowState);

        AnalyzeMethod(isBrief, ID2D1HwndRenderTarget_VTable, 58, "ID2D1HwndRenderTarget", "Resize",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1HwndRenderTarget::*)(const D2D1_SIZE_U *)>(&ID2D1HwndRenderTarget::Resize));

        AnalyzeMethod_StdCall_const(isBrief, ID2D1HwndRenderTarget_VTable, 59, "ID2D1HwndRenderTarget", "GetHwnd", &ID2D1HwndRenderTarget::GetHwnd);
    }
    std::cout << std::endl;

    // 3. Check methods of ID2D1Bitmap.
    printf("%p %s\n", m_pBitmap, "ID2D1Bitmap");
    void **ID2D1Bitmap_VTable = *(void ***) m_pBitmap;
    {
        AnalyzeMethod(isBrief, ID2D1Bitmap_VTable, 0, "ID2D1Bitmap", "QueryInterface",
                      static_cast<HRESULT(STDMETHODCALLTYPE ID2D1Bitmap::*)(REFIID, void **)>(&ID2D1Bitmap::QueryInterface));

        AnalyzeMethod(isBrief, ID2D1Bitmap_VTable, 1, "ID2D1Bitmap", "AddRef", &ID2D1Bitmap::AddRef);

        AnalyzeMethod(isBrief, ID2D1Bitmap_VTable, 2, "ID2D1Bitmap", "Release", &ID2D1Bitmap::Release);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1Bitmap_VTable, 3, "ID2D1Bitmap", "GetFactory", &ID2D1Bitmap::GetFactory);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1Bitmap_VTable, 4, "ID2D1Bitmap", "GetSize", &ID2D1Bitmap::GetSize);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1Bitmap_VTable, 5, "ID2D1Bitmap", "GetPixelSize", &ID2D1Bitmap::GetPixelSize);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1Bitmap_VTable, 6, "ID2D1Bitmap", "GetPixelFormat", &ID2D1Bitmap::GetPixelFormat);

        AnalyzeMethod_StdCall_const(isBrief, ID2D1Bitmap_VTable, 7, "ID2D1Bitmap", "GetDpi", &ID2D1Bitmap::GetDpi);

        AnalyzeMethod(isBrief, ID2D1Bitmap_VTable, 8, "ID2D1Bitmap", "CopyFromBitmap", &ID2D1Bitmap::CopyFromBitmap);

        AnalyzeMethod(isBrief, ID2D1Bitmap_VTable, 9, "ID2D1Bitmap", "CopyFromRenderTarget", &ID2D1Bitmap::CopyFromRenderTarget);

        AnalyzeMethod(isBrief, ID2D1Bitmap_VTable, 10, "ID2D1Bitmap", "CopyFromMemory", &ID2D1Bitmap::CopyFromMemory);
    }
    std::cout << std::endl;

    // 4. Check methods of IWICImagingFactory.
    IWICImagingFactory *pIWICFactory = nullptr;
    HRESULT hr = CoCreateInstance(
        CLSID_WICImagingFactory, nullptr, CLSCTX_INPROC_SERVER,
        IID_PPV_ARGS(&pIWICFactory)
    );
    if (FAILED(hr)) {
        return;
    }

    printf("%p %s\n", pIWICFactory, "IWICImagingFactory");
    void **IWICImagingFactory_VTable = *(void ***) pIWICFactory;
    {
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 0, "IWICImagingFactory", "QueryInterface",
                      static_cast<HRESULT(STDMETHODCALLTYPE IWICImagingFactory::*)(REFIID, void **)>(&IWICImagingFactory::QueryInterface));

        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 1, "IWICImagingFactory", "AddRef", &IWICImagingFactory::AddRef);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 2, "IWICImagingFactory", "Release", &IWICImagingFactory::Release);

        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 3, "IWICImagingFactory", "CreateDecoderFromFilename", &IWICImagingFactory::CreateDecoderFromFilename);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 4, "IWICImagingFactory", "CreateDecoderFromStream", &IWICImagingFactory::CreateDecoderFromStream);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 5, "IWICImagingFactory", "CreateDecoderFromFileHandle", &IWICImagingFactory::CreateDecoderFromFileHandle);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 6, "IWICImagingFactory", "CreateComponentInfo", &IWICImagingFactory::CreateComponentInfo);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 7, "IWICImagingFactory", "CreateDecoder", &IWICImagingFactory::CreateDecoder);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 8, "IWICImagingFactory", "CreateEncoder", &IWICImagingFactory::CreateEncoder);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 9, "IWICImagingFactory", "CreatePalette", &IWICImagingFactory::CreatePalette);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 10, "IWICImagingFactory", "CreateFormatConverter", &IWICImagingFactory::CreateFormatConverter);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 11, "IWICImagingFactory", "CreateBitmapScaler", &IWICImagingFactory::CreateBitmapScaler);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 12, "IWICImagingFactory", "CreateBitmapClipper", &IWICImagingFactory::CreateBitmapClipper);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 13, "IWICImagingFactory", "CreateBitmapFlipRotator", &IWICImagingFactory::CreateBitmapFlipRotator);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 14, "IWICImagingFactory", "CreateStream", &IWICImagingFactory::CreateStream);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 15, "IWICImagingFactory", "CreateColorContext", &IWICImagingFactory::CreateColorContext);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 16, "IWICImagingFactory", "CreateColorTransformer", &IWICImagingFactory::CreateColorTransformer);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 17, "IWICImagingFactory", "CreateBitmap", &IWICImagingFactory::CreateBitmap);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 18, "IWICImagingFactory", "CreateBitmapFromSource", &IWICImagingFactory::CreateBitmapFromSource);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 19, "IWICImagingFactory", "CreateBitmapFromSourceRect", &IWICImagingFactory::CreateBitmapFromSourceRect);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 20, "IWICImagingFactory", "CreateBitmapFromMemory", &IWICImagingFactory::CreateBitmapFromMemory);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 21, "IWICImagingFactory", "CreateBitmapFromHBITMAP", &IWICImagingFactory::CreateBitmapFromHBITMAP);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 22, "IWICImagingFactory", "CreateBitmapFromHICON", &IWICImagingFactory::CreateBitmapFromHICON);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 23, "IWICImagingFactory", "CreateComponentEnumerator", &IWICImagingFactory::CreateComponentEnumerator);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 24, "IWICImagingFactory", "CreateFastMetadataEncoderFromDecoder", &IWICImagingFactory::CreateFastMetadataEncoderFromDecoder);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 25, "IWICImagingFactory", "CreateFastMetadataEncoderFromFrameDecode", &IWICImagingFactory::CreateFastMetadataEncoderFromFrameDecode);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 26, "IWICImagingFactory", "CreateQueryWriter", &IWICImagingFactory::CreateQueryWriter);
        AnalyzeMethod(isBrief, IWICImagingFactory_VTable, 27, "IWICImagingFactory", "CreateQueryWriterFromReader", &IWICImagingFactory::CreateQueryWriterFromReader);
    }
    std::cout << std::endl;

    ULONG refCount = pIWICFactory->Release();
    if (refCount != 0) {
        return;
    }
}
