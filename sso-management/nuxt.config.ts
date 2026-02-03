// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
    devtools: { enabled: true },

    modules: [
        '@nuxtjs/tailwindcss',
        '@pinia/nuxt'
    ],

    app: {
        head: {
            title: 'SSO Management Dashboard',
            meta: [
                { charset: 'utf-8' },
                { name: 'viewport', content: 'width=device-width, initial-scale=1' },
                { name: 'description', content: 'SSO System Management Dashboard' }
            ]
        }
    },

    components: [
        { path: '~/components/ui', pathPrefix: false },
        '~/components'
    ],

    runtimeConfig: {
        public: {
            apiBase: process.env.API_BASE_URL || 'http://localhost:3000'
        }
    },

    tailwindcss: {
        cssPath: '~/assets/css/main.css',
        configPath: 'tailwind.config.js'
    },

    ssr: false, // SPA mode for admin dashboard

    compatibilityDate: '2024-01-27',

    devServer: {
        port: 3002
    }
})
