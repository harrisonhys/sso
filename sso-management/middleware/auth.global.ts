export default defineNuxtRouteMiddleware((to) => {
    // Skip login page
    if (to.path === '/login') return

    if (process.client) {
        const user = localStorage.getItem('user')
        if (!user) {
            return navigateTo('/login')
        }
    }
})
