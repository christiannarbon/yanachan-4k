import { createApp } from 'vue'
import App from './App.vue'
import './styles/theme.css'
import './styles/art-themes.css'
import './style.css'
// Last: the Yanami theme's own drawing, which overrides component rules.
import './styles/calorie-meter.css'

createApp(App).mount('#app')
