# eDonish Auto v6.0.0

## Что нового

### Новая архитектура internal/ui/
- Полностью переписанный UI с чистой архитектурой `internal/ui/`
- Пакет `internal/config/` — централизованная конфигурация
- Пакет `internal/edonish/` — чистый API-клиент без зависимостей от UI
- Новые компоненты: LoginPage, Dashboard, GradesTab, ScheduleTab, HomeworkTab, DiariesTab

### Исправления компиляции
- `fyne.Color` / `fyne.NewColor` заменены на `image/color.NRGBA` (совместимость с Fyne v2.7)
- `theme.ColorIcon()` заменён на `theme.ColorPaletteIcon()` (Fyne v2.7 API)
- Удалены неиспользуемые импорты и переменные

### Улучшения
- Сохранение сессии (логин/пароль/школа) в `~/.edonish_session.json`
- Переключение тёмной/светлой темы
- Автовыбор школы если она одна
- Автообновление JWT-токена при 401

## Доступные файлы

| Платформа | Файл | Описание |
|-----------|------|----------|
| Linux | `edonish-app-linux` | Прямой запуск |
| Linux DEB | `edonish-app_v6.0.0_amd64.deb` | Для Ubuntu/Debian |
| Linux RPM | `edonish-app-v6.0.0-1.x86_64.rpm` | Для Fedora/RHEL |
| Windows | `edonish-app-windows.exe` | Прямой запуск |

## Установка

### Linux (DEB)
```bash
sudo dpkg -i edonish-app_v6.0.0_amd64.deb
sudo apt-get install -f
edonish-app
```

### Linux (RPM)
```bash
sudo rpm -i edonish-app-v6.0.0-1.x86_64.rpm
edonish-app
```

### Windows
Запустите `edonish-app-windows.exe`
