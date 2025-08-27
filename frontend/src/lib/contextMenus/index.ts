// src/lib/contextMenus/index.ts

// Tüm spesifik menü tanımlarını içe aktarın
import { getFlightItemContextMenu } from './flightItem';
import { getTimelineHeaderContextMenu } from './timelineHeader';
import { getTimelineRowContextMenu } from './timelineRow';
import { getTimelineEmptyContextMenu } from './timelineEmpty';
import { getTimelineRowSidebarContextMenu } from './timelineRowSidebar'; // YENİ EKLENDİ

// İhtiyaç duyulan tüm menü fonksiyonlarını dışa aktarın
export {
    getFlightItemContextMenu,
    getTimelineHeaderContextMenu,
    getTimelineRowContextMenu,
    getTimelineEmptyContextMenu,
    getTimelineRowSidebarContextMenu,
};