# Feed Task: Sakypaky katalog

## Metadata

- Supplier: Sakypaky
- Generator name: `sakypaky`
- Output endpoint: `/feeds/sakypaky.xml`
- Source URL: `https://www.sakypaky.cz/export/b2b_partners_cs.xml`
- Priority: druhá fáze
- Status: draft, odloženo mimo první MVP
- Last updated: 2026-05-28

## Cíl

Generovat Shoptet XML feed ze Sakypaky katalogového feedu.

## Aktuální pravidla

- Feed je odložený do druhé fáze.
- Zdroj je pravděpodobně podobný Heureka / Zboží stylu.
- Bude nutná transformace do Shoptet XML.
- Kategorie bude potřeba řešit přes společné mapování.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Bezpečná základní pole:
  - `NAME`
  - `EAN`, pokud je ve zdroji
  - `PRICE_VAT`, pokud je ve zdroji
  - `STOCK` nebo dostupnost, pokud je ve zdroji

## Otevřené otázky

- Jaká je přesná XML struktura Sakypaky feedu?
- Které pole je stabilní produktový kód?
- Obsahuje feed varianty?
- Jak mapovat kategorie na Shoptet?
- Má Sakypaky zakládat nové produkty, nebo jen aktualizovat existující?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier sakypaky` vytvoří validní XML.
- Výstup je dostupný na `/feeds/sakypaky.xml`.
- Výstup nepřekročí Shoptet limity.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné Sakypaky položky.
- Unit test chybějícího produktu kódu.
- Unit test variant, pokud je zdroj obsahuje.
- Unit test kategorie po doplnění mapování.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
