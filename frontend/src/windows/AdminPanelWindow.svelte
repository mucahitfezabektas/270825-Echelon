<script lang="ts">
  import { fade } from "svelte/transition";
  import { quintOut } from "svelte/easing";
  import { onMount, onDestroy } from "svelte";
  import { get } from "svelte/store";
  import { authStore, logout } from "@/stores/authStore";
  import { showAlert } from "@/stores/alertStore"; // Custom alert/toast için

  let selectedFile: File | null = null;
  let periodMonth: string = "";
  let resetTable: boolean = false;
  let isLoading: boolean = false;
  let uploadProgress: number = 0;
  let availablePeriods: string[] = [];
  let showNewPeriodInput: boolean = false;

  // YÜKLEME TİPİ GÜNCELLENDİ: "crew_info" eklendi
  let uploadType:
    | "actual"
    | "publish"
    | "activity_code"
    | "off_day_table"
    | "penalty"
    | "aircraft_crew_need"
    | "crew_document"
    | "crew_info" = "actual"; // ✅ Yeni tip eklendi

  const BACKEND_BASE_URL = "http://localhost:8080";

  let authToken: string | null = null;
  const unsubscribeAuth = authStore.subscribe((state) => {
    authToken = state.token;
  });

  onMount(async () => {
    authToken = get(authStore).token;
    if (!authToken) {
      showAlert(
        "Yetkilendirme bilgisi bulunamadı. Lütfen tekrar giriş yapın.",
        "Yetkilendirme Hatası",
        "error",
        5000
      );
      console.error("AdminPanelWindow: Yetkilendirme token'ı mevcut değil.");
    }
  });

  onDestroy(() => {
    unsubscribeAuth();
  });

  async function fetchAvailablePeriods() {
    if (!authToken) {
      return;
    }

    try {
      const response = await fetch(`${BACKEND_BASE_URL}/api/period-range`, {
        headers: { Authorization: `Bearer ${authToken}` },
      });
      if (response.ok) {
        availablePeriods = await response.json();
      } else if (response.status === 401 || response.status === 403) {
        const errorText = await response.text();
        let errorMessage = errorText;
        try {
          const errorData = JSON.parse(errorText);
          errorMessage = errorData.error || errorData.message || errorText;
        } catch (jsonError) {
          console.warn("API yanıtı JSON formatında değil:", errorText);
        }
        showAlert(
          `Yetkilendirme hatası: ${errorMessage}. Lütfen tekrar giriş yapın.`,
          "Yetkilendirme Hatası",
          "error",
          5000
        );
        logout();
      } else {
        const errorText = await response.text();
        let errorMessage = errorText;
        try {
          const errorData = JSON.parse(errorText);
          errorMessage = errorData.error || errorData.message || errorText;
        } catch (jsonError) {
          console.warn("API yanıtı JSON formatında değil:", errorText);
        }
        showAlert(
          `Dönem ayları yüklenirken hata oluştu: ${errorMessage}`,
          "Veri Hatası",
          "error",
          5000
        );
      }
    } catch (error: any) {
      showAlert(
        `Dönem ayları sunucudan alınamadı. Ağ bağlantınızı kontrol edin: ${error.message}`,
        "Ağ Hatası",
        "error",
        5000
      );
    }
  }

  function handleFileChange(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      selectedFile = input.files[0];
    } else {
      selectedFile = null;
    }
  }

  async function handleUpload() {
    if (!authToken) {
      showAlert(
        "Dosya yüklemek için yetkilendirme token'ı gerekli. Lütfen tekrar giriş yapın.",
        "Yetkilendirme Hatası",
        "error",
        5000
      );
      return;
    }

    if (!selectedFile) {
      showAlert(
        "Lütfen yüklenecek bir dosya seçin.",
        "Dosya Eksik",
        "error",
        3000
      );
      return;
    }

    let backendUrl: string;
    let fileFieldName: string;

    if (uploadType === "activity_code") {
      const fileName = selectedFile.name;
      const fileExtension = fileName
        .substring(fileName.lastIndexOf("."))
        .toLowerCase();

      if (fileExtension !== ".csv" && fileExtension !== ".xlsx") {
        showAlert(
          "Aktivite Kodu yüklemek için lütfen bir CSV veya XLSX dosyası seçin.",
          "Yanlış Dosya Tipi",
          "error",
          4000
        );
        return;
      }
      backendUrl = `${BACKEND_BASE_URL}/api/activity-code/import-data`;
      fileFieldName = "file";
    } else if (uploadType === "off_day_table") {
      const fileName = selectedFile.name;
      const fileExtension = fileName
        .substring(fileName.lastIndexOf("."))
        .toLowerCase();

      if (fileExtension !== ".csv" && fileExtension !== ".xlsx") {
        showAlert(
          "Boş Gün Tablosu yüklemek için lütfen bir CSV veya XLSX dosyası seçin.",
          "Yanlış Dosya Tipi",
          "error",
          4000
        );
        return;
      }
      backendUrl = `${BACKEND_BASE_URL}/api/off-day-table/import-data`;
      fileFieldName = "file";
    } else if (uploadType === "penalty") {
      const fileName = selectedFile.name;
      const fileExtension = fileName
        .substring(fileName.lastIndexOf("."))
        .toLowerCase();

      if (fileExtension !== ".csv" && fileExtension !== ".xlsx") {
        showAlert(
          "Ceza Bilgisi yüklemek için lütfen bir CSV veya XLSX dosyası seçin.",
          "Yanlış Dosya Tipi",
          "error",
          4000
        );
        return;
      }
      backendUrl = `${BACKEND_BASE_URL}/api/penalties/import-data`;
      fileFieldName = "file";
    } else if (uploadType === "aircraft_crew_need") {
      const fileName = selectedFile.name;
      const fileExtension = fileName
        .substring(fileName.lastIndexOf("."))
        .toLowerCase();

      if (fileExtension !== ".csv" && fileExtension !== ".xlsx") {
        showAlert(
          "Uçak Ekip İhtiyacı yüklemek için lütfen bir CSV veya XLSX dosyası seçin.",
          "Yanlış Dosya Tipi",
          "error",
          4000
        );
        return;
      }
      backendUrl = `${BACKEND_BASE_URL}/api/aircraft-crew-need/import-data`;
      fileFieldName = "file";
    } else if (uploadType === "crew_document") {
      const fileName = selectedFile.name;
      const fileExtension = fileName
        .substring(fileName.lastIndexOf("."))
        .toLowerCase();

      if (fileExtension !== ".csv" && fileExtension !== ".xlsx") {
        showAlert(
          "Ekip Dokümanı yüklemek için lütfen bir CSV veya XLSX dosyası seçin.",
          "Yanlış Dosya Tipi",
          "error",
          4000
        );
        return;
      }
      backendUrl = `${BACKEND_BASE_URL}/api/crew-documents/import-data`;
      fileFieldName = "file";
    } else if (uploadType === "crew_info") {
      // YENİ: CREW_INFO MANTIĞI
      const fileName = selectedFile.name;
      const fileExtension = fileName
        .substring(fileName.lastIndexOf("."))
        .toLowerCase();

      if (fileExtension !== ".csv" && fileExtension !== ".xlsx") {
        showAlert(
          "Ekip Bilgisi yüklemek için lütfen bir CSV veya XLSX dosyası seçin.",
          "Yanlış Dosya Tipi",
          "error",
          4000
        );
        return;
      }
      backendUrl = `${BACKEND_BASE_URL}/api/crew-info/import-data`; // Backenddeki yeni handler adı
      fileFieldName = "file"; // Backenddeki beklenen form alanı adı
    } else {
      // actual veya publish
      if (!selectedFile.name.endsWith(".xlsx")) {
        showAlert(
          "Actual/Publish veri yüklemek için lütfen bir XLSX dosyası seçin.",
          "Yanlış Dosya Tipi",
          "error",
          4000
        );
        return;
      }
      if (!periodMonth) {
        showAlert(
          "Lütfen dönemi (YYYY-AA) belirtin.",
          "Dönem Eksik",
          "error",
          3000
        );
        return;
      }
      const monthRegex = /^\d{4}-\d{2}$/;
      if (!monthRegex.test(periodMonth)) {
        showAlert(
          "Dönem formatı hatalı. Lütfen YYYY-AA formatını kullanın (örn: 2025-07).",
          "Format Hatası",
          "error",
          4000
        );
        return;
      }
      backendUrl = `${BACKEND_BASE_URL}/api/${uploadType}/import-xlsx?month=${periodMonth}`;
      fileFieldName = `${uploadType}_file_xlsx`;
    }

    if (resetTable) {
      if (uploadType === "actual" || uploadType === "publish") {
        backendUrl += "&reset=true";
      } else if (
        uploadType === "activity_code" ||
        uploadType === "off_day_table" ||
        uploadType === "penalty" ||
        uploadType === "aircraft_crew_need" ||
        uploadType === "crew_document" ||
        uploadType === "crew_info" // CREW_INFO EKLENDİ
      ) {
        backendUrl += backendUrl.includes("?") ? "&reset=true" : "?reset=true";
      }
    }

    isLoading = true;
    uploadProgress = 0;

    let progressInterval = setInterval(() => {
      uploadProgress = Math.min(uploadProgress + 5, 95);
    }, 300);

    showAlert(
      "Dosya yükleniyor ve veriler işleniyor...",
      "Yükleniyor",
      "info",
      5000
    );

    try {
      const formData = new FormData();
      formData.append(fileFieldName, selectedFile);

      const response = await fetch(backendUrl, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
        body: formData,
      });

      if (response.ok) {
        const result = await response.json();
        showAlert(
          result.message ||
            `${getUploadTypeText(uploadType)} dosyası başarıyla yüklendi ve işlendi.`,
          "Yükleme Başarılı",
          "success",
          3000
        );
        selectedFile = null;
        // Sadece actual/publish için dönem temizle
        if (uploadType === "actual" || uploadType === "publish") {
          periodMonth = "";
        }
        resetTable = false;
        showNewPeriodInput = false;
        const fileInput = document.getElementById(
          "xlsxFileInput"
        ) as HTMLInputElement;
        if (fileInput) fileInput.value = "";

        // Sadece 'actual' veya 'publish' yüklemesinde dönemleri yeniden çek
        if (uploadType === "actual" || uploadType === "publish") {
          await fetchAvailablePeriods();
        }
      } else {
        const errorText = await response.text();
        let errorMessage = `Sunucudan hata (${response.status}): ${errorText}`;
        try {
          const errorData = JSON.parse(errorText);
          errorMessage = errorData.error || errorData.message || errorText;
        } catch (jsonError) {
          console.warn(
            "API yanıtı JSON formatında değil, ham metin kullanılıyor:",
            errorText
          );
        }

        if (response.status === 401 || response.status === 403) {
          showAlert(
            `Yetkilendirme hatası: ${errorMessage}. Lütfen tekrar giriş yapın.`,
            "Yetkilendirme Hatası",
            "error",
            5000
          );
          logout();
        } else {
          showAlert(
            `Yükleme hatası: ${errorMessage}`,
            "Yükleme Hatası",
            "error",
            5000
          );
        }
        console.error("Yükleme hatası:", response.status, errorMessage);
      }
    } catch (error: any) {
      showAlert(
        "Sunucuya bağlanılamadı veya bilinmeyen bir hata oluştu: " +
          error.message,
        "Ağ Hatası",
        "error",
        5000
      );
      console.error("Yükleme işlemi sırasında hata:", error);
    } finally {
      clearInterval(progressInterval);
      uploadProgress = 100;
      isLoading = false;
    }
  }

  function handlePeriodSelectChange() {
    if (periodMonth === "new") {
      showNewPeriodInput = true;
      periodMonth = "";
    } else {
      showNewPeriodInput = false;
    }
  }

  // YÜKLEME TİPİ METNİ GÜNCELLENDİ
  function getUploadTypeText(
    type:
      | "actual"
      | "publish"
      | "activity_code"
      | "off_day_table"
      | "penalty"
      | "aircraft_crew_need"
      | "crew_document"
      | "crew_info" // YENİ: crew_info için metin
  ): string {
    switch (type) {
      case "actual":
        return "Actual";
      case "publish":
        return "Publish";
      case "activity_code":
        return "Aktivite Kodu";
      case "off_day_table":
        return "Boş Gün Tablosu";
      case "penalty":
        return "Ceza Bilgisi";
      case "aircraft_crew_need":
        return "Uçak Ekip İhtiyacı";
      case "crew_document":
        return "Ekip Dokümanı";
      case "crew_info": // YENİ: crew_info için metin
        return "Ekip Bilgisi";
      default:
        return "Dosya";
    }
  }

  // YÜKLEME TİPİ DEĞİŞTİĞİNDE UI SIFIRLAMA MANTIĞI GÜNCELLENDİ
  $: if (uploadType) {
    selectedFile = null;
    periodMonth = "";
    resetTable = false;
    showNewPeriodInput = false;
    const fileInput = document.getElementById(
      "xlsxFileInput"
    ) as HTMLInputElement;
    if (fileInput) fileInput.value = "";

    // Sadece 'actual' veya 'publish' seçildiğinde dönemleri yeniden çek
    if (uploadType === "actual" || uploadType === "publish") {
      fetchAvailablePeriods();
    } else {
      availablePeriods = []; // Diğer tiplerde dönem listesini temizle
    }
  }
</script>

<div class="window-content">
  <h3>Veri Yükleme Paneli</h3>

  <div class="upload-form-group">
    <fieldset class="radio-group-fieldset">
      <legend>Yüklenecek Veri Tipi Seçin:</legend>
      <div class="radio-group">
        <input
          type="radio"
          id="uploadTypeActual"
          name="uploadType"
          value="actual"
          bind:group={uploadType}
        />
        <label for="uploadTypeActual">Actual Veri (XLSX)</label>

        <input
          type="radio"
          id="uploadTypePublish"
          name="uploadType"
          value="publish"
          bind:group={uploadType}
        />
        <label for="uploadTypePublish">Publish Veri (XLSX)</label>

        <input
          type="radio"
          id="uploadTypeActivityCode"
          name="uploadType"
          value="activity_code"
          bind:group={uploadType}
        />
        <label for="uploadTypeActivityCode">Aktivite Kodu (CSV/XLSX)</label>

        <input
          type="radio"
          id="uploadTypeOffDayTable"
          name="uploadType"
          value="off_day_table"
          bind:group={uploadType}
        />
        <label for="uploadTypeOffDayTable">Boş Gün Tablosu (CSV/XLSX)</label>

        <input
          type="radio"
          id="uploadTypePenalty"
          name="uploadType"
          value="penalty"
          bind:group={uploadType}
        />
        <label for="uploadTypePenalty">Ceza Bilgisi (CSV/XLSX)</label>

        <input
          type="radio"
          id="uploadTypeAircraftCrewNeed"
          name="uploadType"
          value="aircraft_crew_need"
          bind:group={uploadType}
        />
        <label for="uploadTypeAircraftCrewNeed"
          >Uçak Ekip İhtiyacı (CSV/XLSX)</label
        >

        <input
          type="radio"
          id="uploadTypeCrewDocument"
          name="uploadType"
          value="crew_document"
          bind:group={uploadType}
        />
        <label for="uploadTypeCrewDocument">Ekip Dokümanı (CSV/XLSX)</label>

        <input
          type="radio"
          id="uploadTypeCrewInfo"
          name="uploadType"
          value="crew_info"
          bind:group={uploadType}
        />
        <label for="uploadTypeCrewInfo">Ekip Bilgisi (CSV/XLSX)</label>
      </div>
    </fieldset>
  </div>

  <div class="upload-form-group">
    <label for="xlsxFileInput">
      {#if uploadType === "actual" || uploadType === "publish"}
        XLSX Dosyasını Seçin:
      {:else if uploadType === "activity_code" || uploadType === "off_day_table" || uploadType === "penalty" || uploadType === "aircraft_crew_need" || uploadType === "crew_document" || uploadType === "crew_info"}
        CSV/XLSX Dosyasını Seçin:
      {/if}
    </label>
    <input
      type="file"
      id="xlsxFileInput"
      accept={uploadType === "actual" || uploadType === "publish"
        ? ".xlsx"
        : ".csv,.xlsx"}
      on:change={handleFileChange}
      aria-label={uploadType === "actual" || uploadType === "publish"
        ? "XLSX dosyası seçimi"
        : "CSV veya XLSX dosyası seçimi"}
    />
    {#if selectedFile}
      <p class="selected-file-info">
        Seçilen Dosya: <strong>{selectedFile.name}</strong> ({(
          selectedFile.size /
          1024 /
          1024
        ).toFixed(2)} MB)
      </p>
    {/if}
  </div>

  {#if uploadType === "actual" || uploadType === "publish"}
    <div class="upload-form-group">
      <label for="periodMonthSelect">Dönem Ayı (YYYY-AA):</label>
      <select
        id="periodMonthSelect"
        bind:value={periodMonth}
        on:change={handlePeriodSelectChange}
        aria-label="Dönem ayı seçimi"
      >
        <option value="">-- Ay Seçin --</option>
        {#each availablePeriods as period}
          <option value={period}>{period}</option>
        {/each}
        <option value="new">-- Yeni Ay Ekle --</option>
      </select>
    </div>

    {#if showNewPeriodInput}
      <div
        class="upload-form-group"
        in:fade={{ duration: 200, easing: quintOut }}
      >
        <label for="newPeriodMonthInput">Yeni Dönem Ayı Girin (YYYY-AA):</label>
        <input
          type="text"
          id="newPeriodMonthInput"
          bind:value={periodMonth}
          placeholder="örn: 2025-07"
          aria-label="Yeni dönem ayı girişi"
          maxlength="7"
          on:input={() => {
            if (periodMonth.length === 4 && !periodMonth.includes("-")) {
              periodMonth += "-";
            }
          }}
        />
      </div>
    {/if}
  {/if}

  <div class="upload-form-group checkbox-group">
    <input type="checkbox" id="resetTableCheckbox" bind:checked={resetTable} />
    <label for="resetTableCheckbox"
      >Yüklemeden önce veritabanını <strong>sıfırla</strong> (Dikkat! Tüm veri silinir)</label
    >
  </div>

  <button on:click={handleUpload} disabled={isLoading} class="upload-button">
    {#if isLoading}
      Yükleniyor...
    {:else}
      {getUploadTypeText(uploadType)} Yükle
    {/if}
  </button>

  {#if isLoading}
    <div class="progress-bar-container">
      <div class="progress-bar" style="width: {uploadProgress}%;"></div>
    </div>
  {/if}
</div>

<style>
  .window-content {
    padding: 20px;
    max-width: 500px;
    margin: 0 auto;
    font-family: Arial, sans-serif;
  }

  h3 {
    color: #333;
    margin-bottom: 25px;
    text-align: center;
    font-size: 1.8em;
  }

  .upload-form-group {
    margin-bottom: 15px;
  }

  .upload-form-group label {
    display: block;
    margin-bottom: 8px;
    font-weight: bold;
    color: #555;
  }

  .upload-form-group input[type="file"],
  .upload-form-group input[type="text"],
  .upload-form-group select {
    width: 100%;
    padding: 10px;
    border: 1px solid #ccc;
    border-radius: 5px;
    box-sizing: border-box;
    font-size: 1em;
    background-color: #fff;
  }

  .upload-form-group input[type="file"] {
    padding: 8px 10px;
    background-color: #f9f9f9;
  }

  .selected-file-info {
    margin-top: 10px;
    font-size: 0.9em;
    color: #666;
  }

  .checkbox-group {
    display: flex;
    align-items: center;
    margin-top: 20px;
  }

  .checkbox-group input[type="checkbox"] {
    margin-right: 10px;
    width: 18px;
    height: 18px;
  }

  .checkbox-group label {
    font-weight: normal;
    margin-bottom: 0;
    display: flex;
    align-items: center;
    cursor: pointer;
  }

  .upload-button {
    display: block;
    width: 100%;
    padding: 12px 20px;
    background-color: #4caf50;
    color: white;
    border: none;
    border-radius: 5px;
    font-size: 1.1em;
    font-weight: bold;
    cursor: pointer;
    transition:
      background-color 0.3s ease,
      opacity 0.3s ease;
    margin-top: 25px;
  }

  .upload-button:hover:not(:disabled) {
    background-color: #45a049;
  }

  .upload-button:disabled {
    background-color: #cccccc;
    cursor: not-allowed;
    opacity: 0.7;
  }

  .progress-bar-container {
    width: 100%;
    background-color: #e0e0e0;
    border-radius: 5px;
    margin-top: 15px;
    overflow: hidden;
    height: 10px;
  }

  .progress-bar {
    height: 100%;
    width: 0%;
    background-color: #007bff;
    border-radius: 5px;
    transition: width 0.3s ease-out;
  }

  .radio-group {
    display: flex;
    flex-wrap: wrap;
    gap: 20px;
    margin-top: 5px;
  }

  .radio-group input[type="radio"] {
    margin-right: 5px;
    width: auto;
  }

  .radio-group label {
    font-weight: normal;
    margin-bottom: 0;
    display: flex;
    align-items: center;
    cursor: pointer;
  }

  .radio-group-fieldset {
    border: 1px solid #ddd;
    padding: 15px;
    border-radius: 8px;
    margin-bottom: 20px;
  }

  .radio-group-fieldset legend {
    font-weight: bold;
    font-size: 1.1em;
    color: #333;
    padding: 0 10px;
    margin-left: -10px;
  }
</style>
