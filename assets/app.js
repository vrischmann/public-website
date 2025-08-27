
const theme = {
    value: localStorage.getItem("theme") || (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"),
};

document.documentElement.setAttribute("data-theme", theme.value);

window.addEventListener("DOMContentLoaded", () => {
    const toggle = document.getElementById("theme-toggle");

    toggle.addEventListener("click", () => {
        theme.value = theme.value === "light" ? "dark" : "light";
        localStorage.setItem("theme", theme.value);
        document.documentElement.setAttribute("data-theme", theme.value);
    });
});
