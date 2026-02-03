import { defu } from 'defu'

export const useAdminFetch = async <T>(url: string, options: any = {}) => {
    const config = useRuntimeConfig()

    const defaults = {
        baseURL: config.public.apiBase,
        // CRITICAL: Include cookies in cross-origin requests
        credentials: 'include',
        headers: {
            'Accept': 'application/json'
        },
        // Handle 401s universally
        onResponseError({ response }) {
            if (response.status === 401) {
                const router = useRouter()
                // Clear local storage if needed
                if (process.client) {
                    localStorage.removeItem('user')
                }
                router.push('/login')
            }
        }
    }

    // Merge options
    const params = defu(options, defaults)

    return await useFetch<T>(url, params)
}
