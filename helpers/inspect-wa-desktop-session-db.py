import sqlite3
import sys


def main() -> int:
    if len(sys.argv) < 2:
        print("Usage: python helpers/inspect-wa-desktop-session-db.py <path-to-session.db>")
        return 2

    db_path = sys.argv[1]
    con = sqlite3.connect(db_path)
    cur = con.cursor()

    objects = cur.execute(
        "select name, type, sql from sqlite_master where type in ('table','view') order by name"
    ).fetchall()

    print(f"objects={len(objects)}")

    keys = ("voip", "relay", "turn", "stun", "call")
    hits = []
    for name, typ, sql in objects:
        sql_l = (sql or "").lower()
        if any(k in name.lower() for k in keys) or any(k in sql_l for k in keys):
            hits.append((name, typ))

    print(f"name_hits={len(hits)}")
    for name, typ in hits[:80]:
        print(f"{typ}: {name}")

    # Show a few tables and their columns.
    print("--- sample_tables ---")
    shown = 0
    for name, typ, sql in objects:
        if typ != "table":
            continue
        cols = cur.execute(f"pragma table_info({name})").fetchall()
        colnames = [f"{c[1]}:{c[2] or ''}" for c in cols]
        print(f"{name} cols={len(cols)}")
        print("  " + ", ".join(colnames[:20]))
        if len(colnames) > 20:
            print("  ...")
        shown += 1
        if shown >= 5:
            break

    # Best-effort: try to locate the Desktop attr4024 prefix in any TEXT/BLOB column.
    needle_hex_prefix = "0A0618D4B8E2A9"  # from Desktop 0x4024 in our capture
    found = []

    for name, typ, sql in objects:
        if typ != "table":
            continue
        cols = cur.execute(f"pragma table_info({name})").fetchall()
        for _, cname, ctype, *_ in cols:
            ctype_u = (ctype or "").upper()
            is_blob = "BLOB" in ctype_u
            is_text = any(t in ctype_u for t in ("TEXT", "CHAR", "CLOB"))
            if not is_blob and not is_text:
                continue

            try:
                if is_blob:
                    q = (
                        f"select rowid from {name} where {cname} is not null "
                        f"and instr(hex({cname}), ?) > 0 limit 1"
                    )
                    r = cur.execute(q, (needle_hex_prefix,)).fetchone()
                else:
                    q = (
                        f"select rowid from {name} where {cname} is not null "
                        f"and instr(upper({cname}), ?) > 0 limit 1"
                    )
                    r = cur.execute(q, (needle_hex_prefix,)).fetchone()

                if r:
                    found.append((name, cname, r[0]))
                    print(f"FOUND needle in {name}.{cname} rowid={r[0]}")
                    # Stop early after first hit.
                    raise StopIteration
            except StopIteration:
                break
            except Exception:
                # Ignore query errors for tables with reserved keywords, etc.
                pass

        if found:
            break

    print(f"needle_found={len(found)}")

    con.close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
