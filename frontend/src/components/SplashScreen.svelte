<script lang="ts">
  export let message: string = "Yükleniyor...";
</script>

<div class="splash-container" role="alert" aria-live="polite">
  <div class="logo">miniCMS</div>
  <div class="loader">{message}</div>
</div>

<style>
  .splash-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100vh;
    background-color: var(--thy-light-gray);
    color: var(--thy-text-color);
    font-family: "Inter", sans-serif;
  }

  .logo {
    font-size: 2rem;
    font-weight: bold;
    margin-bottom: 1rem;
  }

  .loader {
    font-size: 1rem;
    animation: blink 1s infinite;
  }

  @keyframes blink {
    0%,
    100% {
      opacity: 0.2;
    }
    50% {
      opacity: 1;
    }
  }
</style>
