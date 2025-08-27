<script lang="ts">
  export let appInfo = [
    { item: "Backend REST API", value: "8080" },
    { item: "AI / NLP App Port", value: "8181" },
    { item: "Socket.IO", value: "3001" },
    { item: "Frontend Port", value: "3000" },
    {
      item: "Autoupdate URL",
      value: "\\avrasya\\THY_BSK_EKIP_PLANLAMA\\ORTAK\\CrewManager\Settings.txt",
    },
    { item: "Version", value: "v0.0.1" },
    { item: "Server Hostname", value: "-" },
    { item: "Ports Domain", value: "-" },
    { item: "Active Role", value: "Admin" },
    { item: "User Name", value: "admin" },
    { item: "Crew Rules Domain Version", value: "5000" },

    { item: "Crew Rules App Version", value: "v0.0.1" },
  ];
</script>

<div class="window-content">
  <table class="info-table">
    <thead>
      <tr>
        <th>Item</th>
        <th>Value</th>
      </tr>
    </thead>
    <tbody>
      {#each appInfo as row}
        <tr>
          <td>{row.item}</td>
          <td>{row.value}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<style>
  .window-content {
    flex-grow: 1;
    padding: 0px;
    overflow: auto;
  }

  .info-table {
    width: 100%;
    border-collapse: collapse;
  }

  .info-table th,
  .info-table td {
    padding: 6px 10px;
    border: 1px solid #e0e0e0;
    text-align: left;
    vertical-align: top;
  }

  .info-table th {
    background-color: #e8e8e8;
    font-weight: bold;
    color: #333;
    border-color: #d0d0d0;
    white-space: nowrap;
  }

  .info-table tr:nth-child(even) {
    background-color: #f5f5f5;
  }

  .info-table tr:hover {
    background-color: #eaf6ff;
  }

  .info-table td:first-child {
    font-weight: 500;
    color: #444;
    width: 35%;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .info-table td:last-child {
    word-break: break-all;
  }
</style>
