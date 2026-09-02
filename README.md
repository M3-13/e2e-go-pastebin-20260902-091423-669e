# go-pastebin

Kleine REST-API in Go, die Pastes im Speicher ablegt, abfragt, auflistet und
löscht. Thread-safe über Mutex, Ablauf über `expires_in_seconds`, saubere
Statuscodes und JSON-Fehler.

## Tech Stack

- **language**: Go
- **framework**: `net/http` (Standardbibliothek, kein externes Framework)
- **testing**: `httptest`

## Installation

Voraussetzung: Go 1.22 oder neuer.

```sh
go mod download
```

## Starten (Dev)

```sh
go run .
```

Der Server lauscht auf Port `8080`.

## Endpunkte

Fehlerantworten verwenden ausschließlich die JSON-Struktur `{"error": "string"}`.

- `POST /pastes` — Body `{content: string, language?: string, expires_in_seconds?: int}`; liefert `201 {id}` (16 Hex-Zeichen).
- `GET /pastes/{id}` — `200 {id, content, language, created_at, expires_at}` (RFC3339); unbekannt/abgelaufen `404`.
- `GET /pastes` — `200 [{id, language, created_at, expires_at}]` ohne `content`, `created_at` absteigend.
- `DELETE /pastes/{id}` — `204`; unbekannt `404`.
- `GET /health` — `200`.

Falsche Methode auf bekanntem Pfad liefert `405`; unbekannter Pfad `404`.

## Tests

```sh
go test ./...
```
