VERDICT: CHANGES_REQUESTED

## Gesamtbewertung

Das Produkt ist in den zentralen Sicherheits- und Datenschutzanforderungen bereits solide umgesetzt: Body-Limit vor dem Einlesen, keine Protokollierung von Paste-Inhalten, generische JSON-Fehlerantworten, Standard-Ablauf von 86400 Sekunden, Thread-Safety und saubere Tests. Vor einer Auslieferung bestehen jedoch noch behebbare DSGVO- und CRA-Lücken. Kein fundamentaler Blocker.

## DSGVO

### 1. Abgelaufene Pastes werden nicht aktiv gelöscht
- **Schweregrad:** hoch
- **Befund:** `internal/store/list.go` filtert abgelaufene Pastes nur aus der Liste, löscht sie aber nicht. `internal/store/get.go` löscht einen abgelaufenen Paste nur beim direkten GET. Ein nie abgerufener, abgelaufener Paste bleibt bis zum Prozessende im Speicher. Das verletzt die Speicherbegrenzung nach Art. 5 Abs. 1 lit. e DSGVO und das Recht auf Löschung nach Art. 17 DSGVO.
- **Abhilfe:** In `internal/store/store.go` eine Methode `DeleteExpired(now time.Time)` ergänzen und in `main.go` eine Hintergrund-Routine starten, z. B. mit `time.Ticker` alle Minute, die abgelaufene Einträge entfernt. Zusätzlich Tests für den Cleanup ergänzen.

### 2. Keine Obergrenze für `expires_in_seconds`
- **Schweregrad:** hoch
- **Befund:** `internal/api/create.go` in `resolveDefaults` akzeptiert jeden positiven `expires_in_seconds`. Sehr große Werte führen potenziell zu einem Überlauf von `time.Duration(expiresInSeconds) * time.Second` (int64-Nanosekunden) und zu fehlerhaftem Ablaufverhalten. Außerdem wird keine maximale Speicherdauer garantiert.
- **Abhilfe:** In `internal/api/create.go` eine Konstante `maxExpirySeconds` einführen, z. B. `86400 * 30` (30 Tage). Werte oberhalb der Grenze mit HTTP 400 ablehnen oder auf das Maximum kappen. Die AC-12-Anforderung bleibt erfüllt, da der Standardwert weiterhin maximal 86400 Sekunden beträgt.

### 3. Kein TLS/HTTPS im Server
- **Schweregrad:** hoch, sofern der Dienst über das Internet erreichbar ist
- **Befund:** `main.go` startet ausschließlich `http.ListenAndServe(":8080", handler)`. Paste-Inhalte können personenbezogene Daten enthalten und würden unverschlüsselt übertragen. Das verletzt Art. 32 DSGVO.
- **Abhilfe:** Einen `http.Server` mit `ListenAndServeTLS` konfigurieren oder im Deployment zwingend eine TLS-Terminierung vorsehen und dies in `README.md` dokumentieren. Wenn der Dienst bewusst nur lokal betrieben wird, dies ebenfalls dokumentieren.

### 4. Kein Autorisierungskonzept für DELETE
- **Schweregrad:** mittel
- **Befund:** `internal/api/delete.go` löscht jede bekannte ID ohne Berechtigungsnachweis. Da `GET /pastes` über `internal/api/list.go` alle IDs öffentlich auflistet, kann jeder alle Pastes löschen. Das betrifft Verfügbarkeit, Integrität und Vertraulichkeit nach Art. 32 DSGVO.
- **Abhilfe:** Falls keine anonyme Löschung gewünscht ist, beim POST ein zufälliges Lösch-Token erzeugen, nur als Hash speichern und DELETE nur mit diesem Token zulassen. Andernfalls die bewusste Produktentscheidung im `README.md` dokumentieren und das Risiko abnehmen.

### 5. Datenschutz-Dokumentation und Rechtsgrundlage nicht sichtbar
- **Schweregrad:** mittel
- **Befund:** Im Code ist keine Rechtsgrundlage, kein Verarbeitungsverzeichnis und kein Löschkonzept erkennbar. Die API verarbeitet potenziell personenbezogene Daten in Form freier Paste-Inhalte.
- **Abhilfe:** Im `README.md` einen Abschnitt „Datenschutz/Rechtsgrundlage“ ergänzen: Zweck, Rechtsgrundlage nach Art. 6 Abs. 1 lit. b DSGVO, Löschkonzept, technische und organisatorische Maßnahmen.

### Positiv DSGVO
- Body-Limit über `http.MaxBytesReader` in `internal/api/create.go`.
- Keine Protokollierung von Paste-Inhalten in `main.go`.
- Generische `{"error": ...}`-Antworten ohne interne Details.
- Standard-Ablauf von 86400 Sekunden und nicht-null `expires_at` erfüllt AC-12.

## EU Cyber Resilience Act (CRA)

### 1. HTTP-Server ohne Timeouts
- **Schweregrad:** mittel
- **Befund:** `main.go` verwendet `http.ListenAndServe` ohne `http.Server`. Dadurch fehlen Read-, Write- und Idle-Timeouts. Der Dienst ist anfällig für langsame Verbindungen oder Ressourcenerschöpfung.
- **Abhilfe:** In `main.go` einen `http.Server` mit `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` und `IdleTimeout` verwenden.

### 2. Keine SBOM und keine dokumentierten Sicherheitseigenschaften
- **Schweregrad:** mittel
- **Befund:** Das sichtbare Repository enthält kein SBOM, keine `SECURITY.md` und keine dokumentierten Sicherheitseigenschaften. Für Produkte mit digitalen Elementen verlangt der CRA entsprechende Nachweise.
- **Abhilfe:** Ein SBOM im CycloneDX- oder SPDX-Format erzeugen und einchecken, z. B. mit Syft oder `go version -m`. Eine `SECURITY.md` mit Support- und Patch-Prozess hinzufügen. Sicherheitseigenschaften im `README.md` beschreiben.

### 3. Keine Begrenzung der Anzahl gespeicherter Pastes
- **Schweregrad:** mittel
- **Befund:** `internal/store/store.go` hat keine Obergrenze für die Anzahl der Pastes. Ein Angreifer kann durch viele POST-Anfragen den Speicher füllen.
- **Abhilfe:** Eine konfigurierbare Maximalzahl einführen, z. B. 10.000 Pastes. Bei Überschreitung entweder HTTP 429/503 zurückgeben oder die ältesten ablaufenden Einträge entfernen.

### 4. `newID()` kann den Prozess durch Panik beenden
- **Schweregrad:** niedrig
- **Befund:** `internal/store/create.go` ruft bei einem Fehler von `crypto/rand.Read` `panic(err)` auf. Ein seltener Zufallsfehler würde den gesamten Dienst beenden.
- **Abhilfe:** Den Fehler als Rückgabewert propagieren und im Handler als generischen 500er-JSON-Fehler ausgeben.

### Positiv CRA
- Verwendung der Go-Standardbibliothek ohne externe Abhängigkeiten.
- Thread-Safety durch `sync.RWMutex`.
- Body-Limit, generische Fehlermeldungen und vorhandene Handler-Tests.

## EU-KI-Verordnung (AI Act)

Nicht anwendbar: Das Produkt enthält keine KI-Funktion. Keine Befunde.

## Pflichttexte und UI

Nicht anwendbar: Reines `go-backend` ohne Endnutzer-UI. Keine Impressums-, Cookie-, Consent- oder Barrierefreiheitspflichten. Falls später eine öffentliche Web-UI ergänzt wird, gelten die entsprechenden Pflichten zusätzlich.

## Barrierefreiheit

Nicht anwendbar: Keine öffentliche Web-UI, sondern eine reine JSON-API. Keine Befunde.