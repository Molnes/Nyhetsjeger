/** @type {import('tailwindcss').Config} */
export default {
        content: [
                './internal/web_server/web/**/*.templ',
        ],
        theme: {
                extend: {
                        colors: {
                                'cblue': '#0085FF',
                                'cindigo': '#5B14F2',
                                'cfuchsia': '#CA1FFF',
                                'clightindigo': '#E3D8F1',
                        },
                        borderRadius: {
                                'button': '0.5rem',
                                'card': '0.5rem',
                                'input': '0.5rem',
                                '1/2': '50%',
                        }
                },
                fontFamily: {
                        body: ['Bitter', 'serif'],
                        sans: ['Noto Sans', 'sans-serif'],
                },
        },
        plugins: [],
};
