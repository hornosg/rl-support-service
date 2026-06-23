# deploy — aportes de plataforma de support-service

Config **propia** del servicio para integrarse a la plataforma del lab, **sin mutar** los archivos
centrales (`infra/kong/kong.yml`, `observability/prometheus/prometheus.yml`). El servicio aporta;
el lab **compone**.

| Archivo | Para |
|---------|------|
| `kong/support-service.yml` | Ruta del gateway: `/support_service` → `mc-support-service:8160` (decK-style) |
| `prometheus/support-service.yml` | Target de scrape (`file_sd`-style): `mc-support-service:9090` |

> El **mecanismo de composición** del lab (decK sync para Kong, `file_sd_configs`/Docker SD para
> Prometheus) está pendiente de definición a nivel plataforma. Hasta entonces estos archivos son la
> fuente de verdad del aporte de este servicio y no se duplican en los archivos centrales.

Logs: no requieren config acá — el contenedor ya lleva los labels `logging=promtail` /
`service_name=support-service` y la agregación los toma sola.
