# Automatická aktualizace feedů přes cron

## Metadata

- Status: implemented
- Last updated: 2026-06-01

## Cíl

Dodavatelské feedy nechceme servírovat jako živou proxy. Náš veřejný endpoint má vracet poslední úspěšně vygenerovaný Shoptet XML soubor z Railway Volume.

Když dodavatel aktualizuje svůj zdrojový feed, projeví se to u nás až po dalším rebuild běhu. Rebuild stáhne aktuální dodavatelský XML, převede ho, zvaliduje a teprve potom atomicky přepíše náš výstupní feed.

Pokud download, transformace nebo validace spadne, poslední validní XML zůstane zachované.

## Implementovaná architektura

Chráněné interní rebuild endpointy:

```text
POST /internal/rebuild/{supplier}
POST /internal/rebuild/all
```

Endpointy musí být chráněné tokenem z environment variable, například:

```text
REBUILD_TOKEN
```

Cron job nezapisuje do vlastního odděleného filesystemu. Jen spustí `/app/rebuild-trigger`, který zavolá hlavní web službu, takže rebuild poběží nad stejným `DATA_DIR` / Railway Volume, ze kterého server servíruje `/feeds/...`.

Cron service variables:

```text
REBUILD_URL=https://arpari-xml-feed-parser-railway-production.up.railway.app/internal/rebuild/all
REBUILD_TOKEN=<stejný-token-jako-na-web-službě>
```

Cron service Start Command:

```text
/app/rebuild-trigger
```

Railway config-as-code soubory:

```text
web service: /railway.web.json
cron service: /railway.cron.json
```

Root `railway.json` nesmí obsahovat sdílený `deploy.startCommand`, protože by se stejný příkaz aplikoval na web i cron službu.

## Doporučená frekvence

Pro katalogové feedy stačí aktualizace každé 2-3 dny.

Aktuální návrh:

```text
stima-products: každé 3 dny v noci
stima-stock: každé 3 dny v noci pro první MVP, později častěji podle dohody
stima-stock-price: každé 3 dny v noci pro první MVP, později podle cenové strategie
```

Aktuální Railway cron výraz pro denní běh ve 04:00 českého letního času:

```text
0 2 * * *
```

Railway cron běží v UTC. To znamená:

```text
04:00 v Česku v letním čase
03:00 v Česku v zimním čase
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

1. Hotovo: přidat `REBUILD_TOKEN`.
2. Hotovo: přidat interní endpoint `POST /internal/rebuild/{supplier}`.
3. Hotovo: přidat interní endpoint `POST /internal/rebuild/all`.
4. Hotovo: sdílet rebuild logiku mezi `cmd/rebuild` a HTTP endpointem.
5. Hotovo: zajistit, že jeden feed fail neukončí celý `rebuild/all`.
6. Hotovo: vytvořit cron službu, která zavolá `POST /internal/rebuild/all`.
7. Hotovo: oddělit Railway config pro web a cron službu.
8. Zbývá průběžně po deployi: ověřit `/status` a veřejné `/feeds/*.xml`.

## Aktuální stav feedů

- `stima-products`: MVP implementováno a lokálně ověřeno proti reálnému zdroji.
- `stima-stock`: MVP implementováno a lokálně ověřeno proti reálnému zdroji.
- `stima-stock-price`: MVP implementováno a lokálně ověřeno proti reálnému zdroji.
- `autronic-products`: MVP implementováno a lokálně ověřeno proti reálnému zdroji.
- `autronic-products-test`: testovací feed s 5 produkty implementován.
- `autronic-availability`: MVP implementováno a lokálně ověřeno proti reálnému zdroji.
- `sego`: MVP implementováno a lokálně ověřeno proti reálnému zdroji.
- `sego-test`: testovací feed s 5 produkty implementován.
- `hon`: MVP implementováno a lokálně ověřeno proti reálnému zdroji.
- `hon-test`: testovací feed s 5 produkty implementován.

## Poznámka

Cílový stav: katalogové a aktualizační feedy se nemají spouštět při startu serveru. Server má jen servírovat poslední publikovaný výstup. Automatické aktualizace má řešit cron.
