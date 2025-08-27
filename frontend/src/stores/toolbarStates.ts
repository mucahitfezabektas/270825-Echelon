import { writable } from 'svelte/store';
import { getWindowConfig, type WindowConfig } from '@/config/windowConfig';

export type SeparatorItem = { id: string; type: "separator" };
export type ToolbarItem = (WindowConfig & { type: "item" }) | SeparatorItem;

// Varsayılan toolbar için ID'ler ve separator tanımları
const defaultToolbarActionIds: (string | SeparatorItem)[] = [
  "OperationWindow",
  "ColorSamplerWindow",
  "UnifiedTableEditorWindow",
  "FilterQueryWindow",
  { id: "sep-1", type: "separator" },
  "CustomizeWindow",
  "UserOptionsWindow",
  "AdminPanelWindow",
  { id: "sep-2", type: "separator" },
  "OpenProjectWindow",
  "SaveChangesWindow",
  { id: "sep-3", type: "separator" },
  "AboutWindow",
  "ExitWindow",
];

function createToolbarActions(items: (string | SeparatorItem)[]): ToolbarItem[] {
  return items.map((item) => {
    if (typeof item === "string") {
      const config = getWindowConfig(item);
      if (!config) {
        console.warn(`[toolbarStates] WindowConfig not found for ID: ${item}`);
        return null;
      }
      return {
        ...config,
        type: "item" as const,
      };
    }
    return item; // SeparatorItem ise doğrudan döndür
  }).filter(Boolean) as ToolbarItem[]; // null olanları filtrele
}

export const toolbarActions = writable<ToolbarItem[]>(createToolbarActions(defaultToolbarActionIds));

export function resetToolbarActions() {
  toolbarActions.set(createToolbarActions(defaultToolbarActionIds));
}

export function updateToolbarActions(newActions: ToolbarItem[]) {
  toolbarActions.set(newActions);
}
