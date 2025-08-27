<script lang="ts">
  export let message: string = "This window content is empty.";
  export let title: string = "Empty Screen";
</script>

<div class="empty-content-wrapper">
  <h3>{title}</h3>
  <p>{message}</p>
</div>

<style>
  .empty-content-wrapper {
    text-align: center;
    color: #555;
    background-color: #f9f9f9;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
  }
  h3 {
    color: #333;
    margin-bottom: 10px;
  }
  p {
    margin-bottom: 5px;
  }
</style>
