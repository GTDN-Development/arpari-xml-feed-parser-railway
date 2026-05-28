# Automatická aktualizace feedů přes cron

## Metadata

- Status: draft, čeká na implementaci
- Last updated: 2026-05-28

## Cíl

Dodavatelské feedy nechceme servírovat jako živou proxy. Náš veřejný endpoint má vracet poslední úspěšně vygenerovaný Shoptet XML soubor z Railway Volume.

Když dodavatel aktualizuje svůj zdrojový feed, projeví se to u nás až po dalším rebuild běhu. Rebuild stáhne aktuální dodavatelský XML, převede ho, zvaliduje a teprve potom atomicky přepíše náš výstupní feed.

Pokud download, transformace nebo validace spadne, poslední validní XML zůstane zachované.

## Navržená architektura

Přidat chráněné interní rebuild endpointy:

```text
POST /internal/rebuild/{supplier}
POST /internal/rebuild/all
```

Endpointy musí být chráněné tokenem z environment variable, například:

```text
REBUILD_TOKEN
```

Cron job pak nebude zapisovat do vlastního odděleného filesystemu. Jen zavolá hlavní web službu, takže rebuild poběží nad stejným `DATA_DIR` / Railway Volume, ze kterého server servíruje `/feeds/...`.

## Doporučená frekvence

Pro katalogové feedy stačí aktualizace každé 2-3 dny.

Aktuální návrh:

```text
stima-products: každé 3 dny v noci
stima-stock: každé 3 dny v noci pro první MVP, později častěji podle dohody
stima-stock-price: každé 3 dny v noci pro první MVP, později podle cenové strategie
```

Railway cron výraz:

```text
0 1 */3 * *
```

Railway cron běží v UTC. To znamená přibližně:

```text
03:00 v Česku v letním čase
02:00 v Česku v zimním čase
```

Pozdější doporučení podle typu feedu:

```text
katalog / produkty: každé 2-3 dny
sklad: 1-4x denně
ceny: 1x denně nebo každé 2-3 dny podle dodavatele
```

## Chování při chybách

- Každý supplier se rebuildí samostatně.
- Chyba jednoho supplieru nesmí zastavit publikování ostatních feedů.
- Neúspěšný rebuild nesmí přepsat poslední validní XML.
- `/status` musí ukazovat poslední výsledek pro každý supplier:
  - `status`,
  - `lastRunAt`,
  - `itemsProcessed`,
  - `itemsSkipped`,
  - `error`.

## Implementační kroky

1. Přidat `REBUILD_TOKEN`.
2. Přidat interní endpoint `POST /internal/rebuild/{supplier}`.
3. Přidat interní endpoint `POST /internal/rebuild/all`.
4. Sdílet rebuild logiku mezi `cmd/rebuild` a HTTP endpointem.
5. Zajistit, že jeden feed fail neukončí celý `rebuild/all`.
6. Na Railway vytvořit cron službu, která každé 3 dny zavolá `POST /internal/rebuild/all`.
7. Ověřit `/status` a veřejné `/feeds/*.xml`.

## Aktuální stav feedů

- `stima-products`: MVP implementováno a lokálně ověřeno proti reálnému zdroji.
- `stima-stock`: MVP implementováno a lokálně ověřeno proti reálnému zdroji.
- `stima-stock-price`: MVP implementováno a lokálně ověřeno proti reálnému zdroji.
- Ostatní dodavatelé jsou zatím draft tasky bez implementace.

## Poznámka

STIMA katalog `stima-products` se nemá spouštět při startu serveru. Server má jen servírovat poslední publikovaný výstup. Automatické aktualizace má řešit cron.
