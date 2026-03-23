// Service Worker для обработки Web Push уведомлений.
// Обрабатывает события push (показ уведомления) и notificationclick (навигация).

self.addEventListener("push", (e) => {
  if (!e.data) return;

  /** @type {{ title: string; body?: string; url?: string }} */
  const data = e.data.json();

  e.waitUntil(
    self.registration.showNotification(data.title, {
      body: data.body || "",
      icon: "/icon-192.png",
      badge: "/icon-192.png",
      data: { url: data.url || "/" },
    })
  );
});

self.addEventListener("notificationclick", (e) => {
  e.notification.close();

  const url = e.notification.data?.url || "/";

  // Пытаемся найти уже открытое окно приложения и активировать его.
  // Если нет — открываем новое.
  e.waitUntil(
    clients
      .matchAll({ type: "window", includeUncontrolled: true })
      .then((clientList) => {
        for (const client of clientList) {
          if (client.url.includes(self.location.origin) && "focus" in client) {
            client.navigate(url);
            return client.focus();
          }
        }
        return clients.openWindow(url);
      })
  );
});
