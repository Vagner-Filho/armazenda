/** @type {import('tailwindcss').Config} */
module.exports = {
    // content: ["./templates/*.html"],
    content: ["./**/*.html", "./**/*.tmpl"],
    theme: {
        extend: {
            boxShadow: {
                'center': '0px 0px 5px 1px rgba(0, 0, 0, 0.3)'
            }
        },
    },
    plugins: [],
}
