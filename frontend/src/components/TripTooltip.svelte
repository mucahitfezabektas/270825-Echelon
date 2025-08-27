<script lang="ts">
  export let tripId: string;
  export let departure_port: string;
  export let arrival_port: string;
  export let departure_time: number;
  export let arrival_time: number;
  export let rotation_minutes: number;
  export let x: number;
  export let y: number;
</script>

<div class="trip-tooltip" style="left: {x}px; top: {y}px;">
  <div><strong>Trip:</strong> {tripId}</div>
  <div>
    <strong>Kalkış:</strong>
    {departure_port} – {new Date(departure_time).toLocaleTimeString()}
  </div>
  <div>
    <strong>Varış:</strong>
    {arrival_port} – {new Date(arrival_time).toLocaleTimeString()}
  </div>
  <div>
    <strong>Süre:</strong>
    {Math.floor(rotation_minutes / 60)} sa {rotation_minutes % 60} dk
  </div>
</div>

<style>
  .trip-tooltip {
    position: absolute;
    background: white;
    border: 1px solid #ccc;
    padding: 6px 8px;
    font-size: 12px;
    border-radius: 4px;
    pointer-events: none;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
    z-index: 100;
  }
</style>
