/**
 * 裝置偵測工具 — 判斷當前平台以決定顯示哪些社交登入選項
 *
 * iOS: Apple + Google
 * Android: Google only
 * Web: Apple + Google
 */

export type Platform = 'ios' | 'android' | 'web'

export function detectPlatform(): Platform {
  const ua = navigator.userAgent || ''

  if (/iPad|iPhone|iPod/.test(ua)) {
    return 'ios'
  }

  if (/Android/.test(ua)) {
    return 'android'
  }

  return 'web'
}

export function shouldShowAppleLogin(): boolean {
  return detectPlatform() !== 'android'
}

export function shouldShowGoogleLogin(): boolean {
  return true // 所有平台都顯示 Google 登入
}
