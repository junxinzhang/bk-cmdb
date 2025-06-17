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
import xid from 'xid-js'

const TRACE_CHARS = 'abcdef0123456789'
const randomString = (length, chars) => {
  let result = ''
  for (let i = length; i > 0; --i) result += chars[Math.random() * chars.length | 0]
  return result
}

// 获取当前用户名
const getCurrentUser = () => {
  // 从Cookie中获取用户名
  const userCookie = document.cookie
    .split(';')
    .find(cookie => cookie.trim().startsWith('user='))
  
  if (userCookie) {
    return userCookie.split('=')[1]
  }
  
  // 兜底使用admin
  return 'admin'
}

export default () => ({
  // opentelementry TraceID
  traceparent: `00-${randomString(32, TRACE_CHARS)}-${randomString(16, TRACE_CHARS)}-01`,
  // 请求ID
  'X-Request-ID': `cc0000${xid.next()}`,
  // 用户认证头 - 解决API认证问题
  'BK_User': getCurrentUser(),
  // 供应商ID头 - 蓝鲸平台必需
  'HTTP_BLUEKING_SUPPLIER_ID': '0'
})
