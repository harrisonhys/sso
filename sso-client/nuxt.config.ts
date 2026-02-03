// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-04-03',
  devtools: { enabled: true },

  modules: [
    '@nuxtjs/tailwindcss',
    '@nuxtjs/color-mode'
  ],

  colorMode: {
    classSuffix: ''
  },

  runtimeConfig: {
    public: {
      ssoServerUrl: process.env.SSO_SERVER_URL || 'http://localhost:8080',
      clientId: process.env.CLIENT_ID || 'demo-client',
      clientSecret: process.env.CLIENT_SECRET || '',
      redirectUri: process.env.REDIRECT_URI || 'http://localhost:3001/callback'
    }
  },

  app: {
    head: {
      title: 'SSO Client Demo',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'SSO Client Demo Application' }
      ],
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }
      ]
    }
  },

  devServer: {
    port: 3001
  },

  ssr: false
})
