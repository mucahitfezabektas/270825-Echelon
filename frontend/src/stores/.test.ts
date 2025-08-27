import { preferencesStore } from "./userOptionsStore";
preferencesStore.subscribe(p => console.log("✏️ Seçilen airport:", p.timezone.selectedAirport));
