/** @type {import('tailwindcss').Config} */
export default {
        content: [
                './internal/web_server/web/**/*.templ',
        ],
        theme: {
                fontFamily: {
                        body: ['Bitter', 'serif'],
                        sans: ['Noto Sans', 'sans-serif'],
                },
                extends: {
                }
        },
        plugins: [],
};
