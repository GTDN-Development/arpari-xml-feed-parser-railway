# Feed Task: Autronic dostupnost

## Metadata

- Supplier: Autronic
- Generator name: `autronic-availability`
- Output endpoint: `/feeds/autronic-availability.xml`
- Source URL: `https://autronic.cz/feeds/availability-feed.xml`
- Priority: první fáze
- Status: MVP implementováno, čeká na produkční deploy a importní ladění
- Last updated: 2026-06-04

## Cíl

Generovat Shoptet XML feed pro aktualizaci dostupnosti, skladu a ceny Autronic produktů.

## Aktuální pravidla

- Zdroj je dostupnostní feed, ne kompletní produktový katalog.
- Parser musí používat `GET`; `HEAD` u tohoto zdroje vrací 404.
- Dostupnostní feed neobsahuje kategorii, proto se sám o sobě nepoužívá jako filtr.
- Parser stahuje také katalogový `product-feed.xml`, používá existující katalogovou transformaci jako filtr a zachovává stejný jednoduchý/variantní tvar jako `autronic-products`.
- Do výstupu se dostanou pouze kódy, které jsou součástí filtrovaného Autronic katalogu.
- Feed nepřepisuje katalogová pole kromě ceny.
- Cenu bere z katalogového feedu stejně jako `autronic-products`: přednostně `RetailPromotionalPriceIncludingVat`, jinak `RetailPriceIncludingVat`.
- U variant se posílá původní variantní `PARAMETERS`, protože Shoptet supplier XML schéma ho u `VARIANTS/VARIANT` vyžaduje.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Dostupnost / sklad:
  - `STOCK`
- Cena:
  - `PRICE_VAT`
- Neposílat:
  - názvy
  - popisy
  - obrázky
  - kategorie
  - produktové parametry kromě povinného variantního parametru u variant

## Otevřené otázky

- Zatím nemapujeme `AvailabilityStatus` na Shoptet textovou dostupnost. Update posílá pouze sklad.

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier autronic-availability` vytvoří validní XML.
- Výstup je dostupný na `/feeds/autronic-availability.xml`.
- Parser používá `GET`, ne `HEAD`.
- Do výstupu se dostanou pouze produkty z filtrovaného Autronic katalogu.
- Výstup projde Shoptet `products-supplier-v10.rng` validací.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test filtrování přes katalogový feed.
- Unit test variantního skladového update včetně povinných variantních parametrů.
- Unit test zdroje bez skladového payloadu.
- Unit test chybějícího `CODE`.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
