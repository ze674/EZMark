/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./templates/**/*.html",
        "./internal/view/**/*.templ",
        "./internal/view/**/*_templ.go" // Это важно для обнаружения классов в сгенерированных Go-файлах
    ],
    theme: {
        extend: {},
    },
    plugins: [],
}