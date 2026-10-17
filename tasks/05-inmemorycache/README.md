# In-memory cache

## Условие

Реализовать пакет `inmemorycache` — простой in-memory кэш с методами:

- `Set(key string, value any)` — записать значение.
- `Get(key string) (any, bool)` — получить значение и флаг наличия.
- `Delete(key string)` — удалить по ключу.

