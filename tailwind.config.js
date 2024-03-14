/** @type {import('tailwindcss').Config} */
export default {
        content: [
                './internal/api/web/**/*.templ',
        ],
        theme: {
                colors: {
                },
                fontFamily: {
                        sans: ['Bitter', 'serif'],
                },
                extends: {
                        fontFamily: {

                                'noto-sans': ['Noto Sans', 'sans-serif'],
                        }
                }
        },
        plugins: [],
};
