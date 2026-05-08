// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { i18n } from './locales'
import TDesign from 'tdesign-vue-next'
import 'tdesign-vue-next/es/style/index.css'
import './styles/index.css'
import './styles/auth.css'
import Icon from './components/Icon.vue'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(TDesign)
app.component('Icon', Icon)

app.mount('#app')
