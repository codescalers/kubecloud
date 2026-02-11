export default defineNuxtPlugin(() => {
  // const api = useApi()
  // console.log("plugin?")

  const { apiBasePath } = useRuntimeConfig().public
  const { accessToken } = useTokens()

  //   const toast = useToast()

  const sse = new EventSource(`${apiBasePath}/events?token=${accessToken.value}`, { withCredentials: true })

  const toast = useToast()
  sse.onopen = () => {
    // console.log("sse opened")
  }

  sse.onmessage = (e) => {
    // console.log("Message", e.data)
    try {
      const d = JSON.parse(e.data)
      const { data } = d
      // console.log(d)

      const fn = (toast as any)[d.severity]
      if (!fn) {
        // console.log(d.severity, "not found")

        return
      }

      fn.bind(toast)({ message: data.message, title: data.subject })
    }
    catch {
      console.error("Failed to parse message", e.data)
    }
  }

  sse.onerror = (e) => {
    console.error("sse error", e)
  }

  //   sse.onmessage = (event) => {
  //     console.log(event)
  //   }

//   sse.onerror = (event) => {
//     console.error(event)
//   }
})
