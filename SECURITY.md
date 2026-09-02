VERDICT: CHANGES_REQUESTED

## Sicherheitsreport

### Zusammenfassung
Die API setzt zentrale Sicherheitsanforderungen der Sprint-Spec um: Body-Limit vor dem vollständigen Einlesen, JSON-Fehler ohne interne Details, keine Protokollierung von Paste-Inhalten, zufällige 16-stellige IDs aus `crypto/rand` und thread-sichere Zugriffe. Es wurden keine kritischen oder hohen Schwachstellen wie Secrets, Injection, Auth-Bypass oder PII-Leaks gefunden. Es bestehen jedoch mittlere und niedrige Härtungslücken: unbegrenzte Speicherbelegung ohne Cleanup, fehlende Obergrenze für `expires_in_seconds`, nicht strikte JSON-Dekodierung sowie fehlende Transport-/Browser-Absicherungsmaßnahmen.

---

### Befunde

#### 1. Medium – Unbegrenzte Speicherbelegung und fehlende Bereinigung abgelaufener Pastes
**Betroffene Stellen:** `internal/store/store.go`, `internal/store/create.go`, `internal/store/list.go`, `internal/store/get.go`, `main.go`

**Beschreibung:**  
`Store.Create` fügt ohne Limit in `map[string]Paste` ein. Es gibt weder eine Ratenbegrenzung noch eine Obergrenze für die Anzahl der gespeicherten Pastes. Abgelaufene Einträge werden in `List` nur übersprungen, aber nicht aus der Map gelöscht; `Get` löscht nur bei direktem Zugriff. Ein Angreifer kann mit vielen POST-Requests (jeweils bis zu 1 MiB Body) den Speicher füllen und den Prozess zum Absturz bringen (DoS). Zusätzlich bleiben abgelaufene Pastes ohne Zugriff dauerhaft im Speicher.

**Fix:**
- In `Store` eine maximale Anzahl `maxPastes` einführen und beim `Create` prüfen; bei Überschreitung älteste abgelaufene oder älteste Einträge entfernen oder mit `507 Insufficient Storage`/`503 Service Unavailable` antworten.
- Periodisches Cleanup (z. B. `time.Ticker`) in `NewStore` oder in `main` starten, das abgelaufene Pastes aus der Map löscht.
- Optional eine Middleware zur Ratenbegrenzung pro Client-IP vor `CreateHandler` schalten.

---

#### 2. Medium – Fehlende Obergrenze/Überlauf bei `expires_in_seconds`
**Betroffene Stellen:** `internal/api/create.go`, `internal/store/create.go`

**Beschreibung:**  
Es wird lediglich ein negativer `expires_in_seconds`-Wert abgelehnt. Ein sehr hoher positiver Wert wird ungeprüft an `time.Duration(expiresInSeconds) * time.Second` übergeben. `time.Duration` ist ein `int64`-Nanosekundenwert; die Multiplikation kann bei großen Werten überlaufen und einen negativen Ablaufzeitpunkt erzeugen. Dadurch ist der Paste sofort abgelaufen oder bleibt bei „nur“ großen Werten über Jahre im Speicher. Dies erleichtert Speicher-DoS und verletzt potenziell die Erwartung, dass ein erstellter Paste abrufbar ist.

**Fix:**
- Eine Konstante `maxExpirySeconds` einführen (z. B. `31536000` = 1 Jahr).
- In `CreateHandler` vor `resolveDefaults`/`Create` prüfen:
  ```go
  if req.ExpiresInSeconds != nil && *req.ExpiresInSeconds > maxExpirySeconds {
      WriteError(w, http.StatusBadRequest, "expires_in_seconds too large")
      return
  }
  ```
- Alternativ in `store.Create` vor der Multiplikation den Wertebereich sicherstellen.

---

#### 3. Low – JSON-Dekodierung akzeptiert weitere Elemente im Body
**Betroffene Stelle:** `internal/api/create.go`

**Beschreibung:**  
`CreateHandler` ruft nur einmal `json.NewDecoder(r.Body).Decode(&req)` auf. Nachfolgende Daten nach dem ersten JSON-Objekt werden ignoriert. Ein Body wie `{"content":"x"} {"content":"y"}` wird als gültig akzeptiert. Das ist kein direkter Exploit, entspricht aber nicht vollständig AC-06 (ungültiges JSON sollte 400 liefern).

**Fix:**
Nach erfolgreichem Decode prüfen, ob weitere Elemente vorhanden sind:
```go
dec := json.NewDecoder(r.Body)
if err := dec.Decode(&req); err != nil { ... }

if dec.More() {
    WriteError(w, http.StatusBadRequest, "invalid JSON")
    return
}
```
Alternativ einen zweiten Decode in eine leere Struktur ausführen und prüfen, ob der Fehler `io.EOF` ist.

---

#### 4. Low – Fehlende Sicherheitsheader für JSON-Antworten
**Betroffene Stelle:** `internal/api/errors.go`

**Beschreibung:**  
`WriteJSON` setzt nur `Content-Type: application/json`. Es fehlt `X-Content-Type-Options: nosniff`. Dieser Header verhindert, dass Browser den JSON-Body als HTML interpretieren, falls die API direkt im Browser aufgerufen wird.

**Fix:**
In `WriteJSON` vor `WriteHeader` ergänzen:
```go
w.Header().Set("X-Content-Type-Options", "nosniff")
```
Das verhindert MIME-Sniffing und bricht keine produktiv genutzte Ressource, da es sich um reine JSON-Antworten handelt.

---

#### 5. Low – Unverschlüsselter Transport auf allen Interfaces
**Betroffene Stelle:** `main.go`

**Beschreibung:**  
`http.ListenAndServe(":8080", handler)` startet den Server unverschlüsselt auf allen Interfaces. Paste-Inhalte können im Klartext über das Netz übertragen werden. Sofern kein TLS-terminierender Reverse Proxy vorgeschaltet ist, ist dies ein Risiko für die Vertraulichkeit.

**Fix:**
- Entweder `http.ListenAndServeTLS` mit Zertifikaten verwenden oder den Dienst hinter einem TLS-Proxy bereitstellen.
- Mindestens in der Betriebsdokumentation festhalten, dass die reine HTTP-Variante nur für isolierte Testumgebungen vorgesehen ist.

---

### Positiv hervorgehoben
- AC-08: `http.MaxBytesReader` begrenzt den Request-Body auf 1 MiB, bevor er vollständig eingelesen wird.
- AC-09/11: Fehlerantworten ausschließlich als `{error: string}`, keine internen Fehlerdetails oder Stacktraces.
- AC-10: Keine Protokollierung von Paste-Inhalten; lediglich eine Startmeldung.
- Zufällige IDs aus `crypto/rand` (16 Hex-Zeichen) statt vorhersagbarer IDs.
- Thread-sichere Map-Zugriffe durch `sync.RWMutex`.
- `GET /pastes` liefert ausschließlich Metadaten ohne `content`.

---

### Dependencies / Scanner
- Es wurde kein anwendbarer Security-Scanner ausgeführt (`(no applicable security scanners for this project type)`); daher liegen keine automatisierten Befunde vor.
- Im sichtbaren Code werden nur Go-Standardbibliotheken verwendet; externe Abhängigkeiten sind nicht erkennbar. Aufgrund der fehlenden Scannerausgabe kann zu bekannten CVEs in der Go-Standardbibliothek keine abschließende Aussage getroffen werden.

---

### Gesamtbewertung
Keine kritischen oder hohen Schwachstellen. Die mittleren Befunde (Speicher-DoS durch unbegrenzte Speicherbelegung, fehlende Ablauf-Obergrenze) sollten vor einem Produktionseinsatz behoben werden. Daher: **CHANGES_REQUESTED**.