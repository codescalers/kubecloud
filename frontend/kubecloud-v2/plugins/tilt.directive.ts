export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.directive("tilt-effect", {
    mounted(el: HTMLElement) {
      el.onmousemove = (e: MouseEvent) => {
        const rect = el.getBoundingClientRect()
        const x = e.clientX - rect.left
        const y = e.clientY - rect.top

        const centerX = rect.width / 2
        const centerY = rect.height / 2

        const rotateX = (y - centerY) / 10
        const rotateY = (centerX - x) / 10

        el.style.transform = `perspective(1000px) rotateX(${rotateX}deg) rotateY(${rotateY}deg) translateY(-10px)`
      }

      el.onmouseleave = () => {
        el.style.transform = `perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(0px)`
      }
    },
    unmounted(el: HTMLElement) {
      el.onmousemove = null
      el.onmouseleave = null
    },
  })
})
