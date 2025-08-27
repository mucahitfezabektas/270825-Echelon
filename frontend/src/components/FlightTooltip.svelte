<script lang="ts">
  export let flight;
  export let x: number;
  export let y: number;
  export let locale: string = "tr-TR";
</script>

<div class="tooltip" style="left: {x}px; top: {y}px;">
  <strong>{flight.activity_code || "N/A"}</strong><br />
  Personel ID: {flight.person_id || "N/A"}<br />
  Trip ID: {flight.trip_id || "N/A"}<br />
  Kuyruk Adı: {flight.plane_tail_name || "N/A"}<br />
  Kalkış: {new Date(flight.departure_time).toLocaleString(locale, {
    dateStyle: "short",
    timeStyle: "short",
  })}<br />
  Varış: {new Date(flight.arrival_time).toLocaleString(locale, {
    dateStyle: "short",
    timeStyle: "short",
  })}<br />
  Kalkış/Varış Portu: {flight.departure_port || "N/A"} → {flight.arrival_port ||
    "N/A"}
</div>

<style>
  .tooltip {
    position: absolute;
    background-color: #333;
    color: white;
    padding: 8px 12px;
    border-radius: 4px;
    font-size: 12px;
    pointer-events: none;
    z-index: 100;
    white-space: nowrap;
    box-shadow: 0 2px 5px rgba(90, 86, 86, 0.2);
  }
</style>
