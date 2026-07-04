# Feed Task: Dřevočal katalog

## Metadata

- Supplier: Dřevočal
- Generator name: `drevocal`
- Output endpoint: `/feeds/drevocal.xml`
- Test generator name: `drevocal-test`
- Test output endpoint: `/feeds/drevocal-test.xml`
- Source URL: `https://www.matrace-drevocal.cz/feed-b2b.xml`
- Reference documentation:
  - `reference/drevocal/drevocal-b2b-feed-dokumentace-2026-05.pdf`
  - `reference/drevocal/drevocal-b2b-feed-dokumentace-v1.1-2026-06.pdf`
- Priority: druhá fáze
- Status: MVP implementováno, doplněna podpora dostupnosti a dárku z feedu v1.1
- Last updated: 2026-06-18

## Cíl

Generovat Shoptet XML feed z Dřevočal B2B katalogového feedu.

## Aktuální pravidla

- Feed je odložený do druhé fáze.
- Nový zdroj je B2B variantní feed ve stylu Heureka XML.
- Jedna zdrojová položka `SHOPITEM` odpovídá jedné variantě matrace.
- Varianty jednoho produktu se sdružují přes `ITEMGROUP_ID`.
- Stabilní kód varianty je `ITEM_ID`.
- Variantní parametry jsou `Rozměr`, `Výška` a `Potah`.
- Feed obsahuje cenu s DPH, měnu, EAN, URL, hlavní obrázek a dostupnost `AVAILABILITY`.
- Výstup záměrně neposílá `DESCRIPTION`, aby automatický import Dřevočal nepřepisoval ručně spravované popisy ani při založení nových produktů.
- Feed v1.1 může obsahovat volitelný element `GIFT`; aktuálně jde o text `polštář Lukáš`.
- Feed neobsahuje sklad po kusech.
- Feed aktuálně neobsahuje kategorii; cílová Shoptet kategorie bude nastavena pravidlem na `LOŽNICE > MATRACE` (`ID=1188`).
- Testovací endpoint `drevocal-test` má používat stejná pravidla, ale pouze prvních 5 výstupních produktů.
- Původní veřejný feed `https://www.matrace-drevocal.cz/feed/` byl jednodušší katalog bez variantních skupin a pro Shoptet napojení se dál nepoužívá.

## Ověření zdroje 2026-06-11

- Aktuální B2B feed má 3 773 variant.
- Varianty jsou rozdělené do 57 produktových skupin.
- Největší skupina má 189 variant, tedy je pod Shoptet limitem 512 variant na produkt.
- Všechny položky mají parametry `Rozměr`, `Výška` a `Potah`.
- EAN chybí u 6 variant.
- Dokumentace uvádí očekávaný rozsah cca 8 000-9 000 variant; před ostrým importem je vhodné ověřit u Dřevočalu, jestli je aktuální feed kompletní.

## Ověření zdroje 2026-06-18

- Aktuální B2B feed má 4 574 variant.
- Varianty jsou rozdělené do 70 produktových skupin.
- Všechny položky mají `AVAILABILITY=Skladem`.
- `GIFT` je u 1 252 variant v 20 produktových skupinách.
- Aktuální hodnota `GIFT` je `polštář Lukáš`.
- Dřevočal v `GIFT` neposílá kód existujícího Shoptet produktu, jen text dárku.

## Implementace

- Stav kódu: implementováno v `internal/drevocal/products.go` a `internal/feed/drevocal_products.go`.
- Registry: supplier `drevocal` a `drevocal-test` jsou dostupné přes `cmd/rebuild`.
- Dostupnost z `AVAILABILITY` se posílá na úroveň variant.
- `GIFT` se posílá jako doplňkový parametr parent produktu `Dárek`.
- Skutečný Shoptet dárek přes `GIFTS > CODE` zatím neposíláme, protože zdroj obsahuje jen text dárku, ne kód dárkového produktu.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild ověřen 2026-06-11:
  - `drevocal`: 3 773 variant přečteno, 57 Shoptet produktů emitováno, 3 773 variant emitováno.
  - `drevocal`: největší variantní produkt má 189 variant.
  - `drevocal-test`: 5 Shoptet produktů emitováno, 377 variant emitováno.
  - Výstupní XML je well-formed a publikace proběhla přes storage publisher.

## MVP rozsah

- Identifikace produktu:
  - parent produkt podle `ITEMGROUP_ID`
  - varianta podle `ITEM_ID`
- Bezpečná základní pole:
  - parent `NAME` odvozený ze společného názvu modelu
  - variant `CODE`
  - variant `EAN`, pokud je ve zdroji
  - variant `PRICE_VAT`
  - `CURRENCY`
  - variant `AVAILABILITY`
  - `IMAGES`
  - doplňkový parametr `Dárek`, pokud zdroj obsahuje `GIFT`
  - cílová Shoptet kategorie `LOŽNICE > MATRACE` (`ID=1188`)
- Variantní parametry:
  - `Rozměr`
  - `Výška`
  - `Potah`

## Otevřené otázky

- Má importer odmítat varianty bez EAN, nebo je importovat bez EAN?
- Má parent produkt používat popis a obrázek z první varianty ve skupině?
- Je aktuální B2B feed kompletní, když dokumentace uvádí vyšší očekávaný počet variant?
- Pokud má být dárek v Shoptetu veden jako skutečný dárkový produkt, je potřeba dodat kód produktu dárku v Shoptetu.

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier drevocal` vytvoří validní XML.
- Výstup je dostupný na `/feeds/drevocal.xml`.
- `go run ./cmd/rebuild --supplier drevocal-test` vytvoří testovací feed s 5 produkty.
- Testovací výstup je dostupný na `/feeds/drevocal-test.xml`.
- Varianty jsou seskupené podle `ITEMGROUP_ID`.
- Každá varianta používá `ITEM_ID` jako Shoptet `CODE`.
- Každá varianta má parametry `Rozměr`, `Výška` a `Potah`.
- Pokud zdroj obsahuje `AVAILABILITY`, varianta má dostupnost.
- Pokud zdroj obsahuje `GIFT`, parent produkt má doplňkový parametr `Dárek`.
- Varianty matrací nepřekročí Shoptet limit 512 variant na produkt.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné Dřevočal položky.
- Unit test seskupení variant podle `ITEMGROUP_ID`.
- Unit test variantních parametrů `Rozměr`, `Výška`, `Potah`.
- Unit test limitu 512 variant.
- Unit test chybějícího produktu kódu.
- Unit test chybějícího EAN.
- Unit test dostupnosti z `AVAILABILITY`.
- Unit test doplňkového parametru `Dárek` z `GIFT`.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
