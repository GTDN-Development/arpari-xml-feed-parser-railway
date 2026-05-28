# Feed Task: Dřevočal katalog

## Metadata

- Supplier: Dřevočal
- Generator name: `drevocal`
- Output endpoint: `/feeds/drevocal.xml`
- Source URL: `https://www.matrace-drevocal.cz/feed/`
- Priority: druhá fáze
- Status: draft, odloženo mimo první MVP
- Last updated: 2026-05-28

## Cíl

Generovat Shoptet XML feed z Dřevočal katalogového feedu.

## Aktuální pravidla

- Feed je odložený do druhé fáze.
- Zdroj je podle zadání jednodušší XML feed.
- Je nutné ověřit varianty matrací, hlavně výšky a dostupné konfigurace.
- Kategorie bude potřeba řešit přes společné mapování.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Bezpečná základní pole:
  - `NAME`
  - `EAN`, pokud je ve zdroji
  - `PRICE_VAT`, pokud je ve zdroji
  - `STOCK` nebo dostupnost, pokud je ve zdroji
- Varianty matrací, pokud je zdroj obsahuje.

## Otevřené otázky

- Jaká je přesná XML struktura Dřevočal feedu?
- Které pole je stabilní produktový kód?
- Jak jsou zapsané varianty matrací?
- Jak řešit rozměry, výšky a další konfigurace?
- Jak mapovat Dřevočal kategorie na Shoptet?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier drevocal` vytvoří validní XML.
- Výstup je dostupný na `/feeds/drevocal.xml`.
- Varianty matrací nepřekročí Shoptet limit 512 variant na produkt.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné Dřevočal položky.
- Unit test matracové varianty.
- Unit test limitu 512 variant.
- Unit test chybějícího produktu kódu.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
