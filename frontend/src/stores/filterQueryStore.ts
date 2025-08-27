//frontend\src\stores\filterQueryStore.ts
import { writable } from 'svelte/store';
import { v4 as uuidv4 } from 'uuid';
import type { SavedFilter, FilterRow } from '@/lib/types';

/* ► Yardımcı: row-list + logic  →  backend sorgu-string */
function buildQuery(rows: FilterRow[], logic: 'AND' | 'OR' = 'AND'): string {
    if (!rows.length) return '';
    const join = ` ${logic} `;
    return rows
        .map(r => {
            const f = r.field.trim(); if (!f || !r.value.trim()) return null;
            switch (r.operator) {
                case 'LIKE':
                    return `${f} LIKE "%${r.value.trim()}%"`;
                default:
                    return `${f} ${r.operator} "${r.value.trim()}"`;
            }
        })
        .filter(Boolean)
        .join(join);
}

/* ► store */
const initial: SavedFilter[] =
    JSON.parse(localStorage.getItem("savedFilters") || "[]");

export const filterQueryStore = writable<SavedFilter[]>(initial);

filterQueryStore.subscribe((filters) => {
    localStorage.setItem("savedFilters", JSON.stringify(filters));
});


export const filterActions = {
    addFilter: (name = "New Filter"): string => {
        const id = uuidv4();
        filterQueryStore.update(fs => [
            ...fs,
            { id, name, rows: [], logic: "AND", query: "" }
        ]);
        return id;
    },

    updateFilterRows: (id: string, rows: FilterRow[]) =>
        filterQueryStore.update(fs =>
            fs.map(f =>
                f.id === id
                    ? { ...f, rows, query: buildQuery(rows, f.logic ?? "AND") }
                    : f
            )
        ),

    updateFilterName: (id: string, name: string) =>
        filterQueryStore.update(fs =>
            fs.map(f => (f.id === id ? { ...f, name } : f))
        ),

    updateFilterLogic: (id: string, logic: "AND" | "OR") =>
        filterQueryStore.update(fs =>
            fs.map(f =>
                f.id === id
                    ? { ...f, logic, query: buildQuery(f.rows, logic) }
                    : f
            )
        ),

    deleteFilter: (id: string) =>
        filterQueryStore.update(fs => fs.filter(f => f.id !== id)) // ✅ bu satırı ekle
};
