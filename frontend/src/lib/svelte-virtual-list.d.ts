declare module 'svelte-virtual-list' {
    import { SvelteComponentTyped } from 'svelte';

    export interface VirtualListProps<T = any> {
        /** Sanal liste verileri */
        items: T[];

        /** Görünen yükseklik (px) */
        height: number;

        /** Her bir item’ın yüksekliği (px) */
        itemHeight: number;

        /** Görünüm dışı buffer eleman sayısı (isteğe bağlı) */
        overscan?: number;
    }

    // Slotlar varsayılan olarak `item` ve `index` gönderiyor
    export default class VirtualList<T = any> extends SvelteComponentTyped<
        VirtualListProps<T>,
        {},
        {
            item: {
                item: T;
                index: number;
            };
        }
    > { }
}
