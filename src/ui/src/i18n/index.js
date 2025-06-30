/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

import Vue from 'vue'
import VueI18n from 'vue-i18n'
import Cookies from 'js-cookie'
import { jsonp } from '@/api'
import { useSiteConfig } from '@/setup/build-in-vars'
import messages from './lang/messages'
import { LANG_COOKIE_NAME, LANG_KEYS, LANG_SET } from './constants'

Vue.use(VueI18n)

const siteConfig = useSiteConfig()

const langInCookie = Cookies.get(LANG_COOKIE_NAME)
console.log('Language in cookie:', langInCookie)
console.log('Available languages:', LANG_SET)
console.log('All cookies:', document.cookie)

const matchedLang = LANG_SET.find(lang => {
  // Try to match both id and apiLocale values
  const idMatch = lang.id === langInCookie
  const apiLocaleMatch = lang.apiLocale === langInCookie
  const aliasMatch = lang.alias?.includes(langInCookie)
  console.log(`Checking ${lang.id}: idMatch=${idMatch}, apiLocaleMatch=${apiLocaleMatch}, aliasMatch=${aliasMatch}`)
  return idMatch || apiLocaleMatch || aliasMatch
})

console.log('Matched language:', matchedLang)
const locale = matchedLang?.id || LANG_KEYS.ZH_CN
console.log('Final locale:', locale)

// If no cookie was found, set the default language cookie
if (!langInCookie) {
  console.log('No language cookie found, setting default to:', locale)
  const defaultCookieValue = LANG_SET.find(lang => lang.id === locale)?.apiLocale || locale
  try {
    Cookies.set(LANG_COOKIE_NAME, defaultCookieValue, {
      expires: 366,
      path: '/',
      sameSite: 'Lax'
    })
    // Also set via document.cookie as backup
    document.cookie = `${LANG_COOKIE_NAME}=${defaultCookieValue}; path=/; max-age=${366 * 24 * 60 * 60}; SameSite=Lax`
    console.log('Default language cookie set to:', defaultCookieValue)
  } catch (e) {
    console.warn('Failed to set default language cookie:', e)
  }
}

const i18n = new VueI18n({
  locale,
  fallbackLocale: LANG_KEYS.ZH_CN,
  messages,
  missing(locale, path) {
    // eslint-disable-next-line no-underscore-dangle
    const parsedPath = i18n._path.parsePath(path)
    return parsedPath[parsedPath.length - 1]
  }
})

export const changeLocale = async (locale) => {
  console.log('Changing locale to:', locale)
  
  const cookieValue = LANG_SET.find(lang => lang.id === locale)?.apiLocale || locale
  console.log('Setting cookie value:', cookieValue)
  
  // Update the i18n locale immediately
  i18n.locale = locale
  console.log('Updated i18n locale to:', i18n.locale)
  
  // Set cookie with simple, reliable options for localhost
  const cookieOptions = {
    expires: 366,
    path: '/',
    sameSite: 'Lax'
  }
  
  console.log('Cookie options:', cookieOptions)
  
  // Remove existing cookies first
  try {
    Cookies.remove(LANG_COOKIE_NAME)
    Cookies.remove(LANG_COOKIE_NAME, { path: '/' })
    Cookies.remove(LANG_COOKIE_NAME, { path: '' })
  } catch (e) {
    console.warn('Error removing old cookies:', e)
  }
  
  // Set the new cookie
  Cookies.set(LANG_COOKIE_NAME, cookieValue, cookieOptions)
  
  // Verify the cookie was set
  const verifyValue = Cookies.get(LANG_COOKIE_NAME)
  console.log('Cookie verification - set value:', cookieValue, 'read value:', verifyValue)
  
  // Also try to set via document.cookie as backup
  document.cookie = `${LANG_COOKIE_NAME}=${cookieValue}; path=/; max-age=${366 * 24 * 60 * 60}; SameSite=Lax`
  
  if (siteConfig?.componentApiUrl) {
    const url = `${siteConfig.componentApiUrl}/api/c/compapi/v2/usermanage/fe_update_user_language/`
    console.log('Calling language sync API:', url)
    try {
      await jsonp(url, { language: cookieValue })
      console.log('Language sync API call successful')
    } catch (error) {
      console.warn('Failed to sync language preference with backend:', error)
    }
  } else {
    console.log('componentApiUrl not configured, skipping backend language sync')
    console.log('siteConfig:', siteConfig)
  }
  
  // Wait and verify cookie is set before reloading
  let retryCount = 0
  const maxRetries = 10
  
  const verifyCookieAndReload = () => {
    const currentCookie = Cookies.get(LANG_COOKIE_NAME)
    console.log(`Cookie verification attempt ${retryCount + 1}: expecting "${cookieValue}", got "${currentCookie}"`)
    
    if (currentCookie === cookieValue || retryCount >= maxRetries) {
      console.log('Cookie verified, reloading page...')
      // Force a hard reload to avoid any caching issues
      window.location.reload(true)
    } else {
      retryCount++
      // Try setting the cookie again
      document.cookie = `${LANG_COOKIE_NAME}=${cookieValue}; path=/; max-age=${366 * 24 * 60 * 60}; SameSite=Lax`
      setTimeout(verifyCookieAndReload, 50)
    }
  }
  
  setTimeout(verifyCookieAndReload, 100)
}

export const language = locale

export const t = (content, ...rest) => i18n.t(content, ...rest)

export default i18n
