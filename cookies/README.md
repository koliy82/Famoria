# Cookies for yt-dlp (YouTube)

Поместите сюда файл `cookies.txt` в формате Netscape, экспортированный из
браузера (см. инструкцию ниже). Папка монтируется в контейнер как read-write,
чтобы yt-dlp мог и читать, и обновлять cookies.

Файл НЕ коммитится в git (см. `.gitignore`).

## Как создать cookies.txt

1. Установите расширение браузера для экспорта cookies в формате Netscape:
   - Chrome / Edge / Brave: «Get cookies.txt LOCALLY»
   - Firefox: «cookies.txt»
2. Откройте https://www.youtube.com и убедитесь, что залогинены
3. Нажмите на расширение → Export → сохраните файл
4. Переименуйте файл в `cookies.txt` и положите в эту папку:
   ```
   cookies/cookies.txt
   ```
5. В `.env` укажите путь внутри контейнера:
   ```
   YTDLP_COOKIES_FILE=/app/cookies/cookies.txt
   ```

Cookies применяются только к YouTube (см. функцию `isYouTube` в коде).
Рекомендуется использовать отдельный YouTube-аккаунт, не основной.
