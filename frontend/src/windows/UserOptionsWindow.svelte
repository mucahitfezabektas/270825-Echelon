<script lang="ts">
  // preferencesStore'u ve Preferences arayüzünü userOptionsStore.ts dosyasından içe aktarıyoruz.
  import {
    preferencesStore,
    type Preferences,
  } from "@/stores/userOptionsStore";
  import { onDestroy, createEventDispatcher } from "svelte"; // Bileşen yok edildiğinde aboneliği kaldırmak için onDestroy ve olay göndermek için createEventDispatcher'ı içe aktarıyoruz.

  // Varsayılan olarak "general" menü öğesini aktif hale getiriyoruz.
  // activeMenuItem'ın tipini Preferences arayüzünün anahtarlarıyla eşleştiriyoruz.
  let activeMenuItem = $$props.activeMenuItem ?? "general";

  // Mağazadaki mevcut tercihleri almak için mağazaya abone oluyoruz.
  // 'preferences' değişkeni, mağaza değiştiğinde otomatik olarak güncellenecektir.
  let preferences: Preferences;
  const unsubscribe = preferencesStore.subscribe((value) => {
    preferences = value;
  });

  // Bileşen yok edildiğinde mağaza aboneliğini otomatik olarak kaldırıyoruz.
  // Bu, bellek sızıntılarını önlemek için iyi bir uygulamadır.
  onDestroy(() => {
    unsubscribe();
  });

  // Olay göndericiyi başlatıyoruz. Bu, üst bileşenlere özel olaylar göndermemizi sağlar.
  const dispatch = createEventDispatcher();

  // Kullanıcıya geri bildirim göstermek için iç durum değişkenleri
  let showFeedback: boolean = false;
  let feedbackMessage: string = "";
  let feedbackType: "success" | "error" = "success";

  // Yan menüden bir öğe seçildiğinde aktif öğeyi güncelleyen fonksiyon.
  function selectMenuItem(item: keyof Preferences) {
    activeMenuItem = item;
  }

  // Bir tercih değiştiğinde mağazayı güncellemek için genel bir fonksiyon.
  // Bu, iç içe geçmiş nesneleri değişmez bir şekilde güncellemeyi sağlar.
  function updatePreference<
    Category extends keyof Preferences,
    Key extends keyof Preferences[Category],
  >(category: Category, key: Key, value: Preferences[Category][Key]) {
    preferencesStore.update((currentPrefs) => {
      return {
        ...currentPrefs, // Mevcut üst düzey tercihleri koru
        [category]: {
          ...currentPrefs[category], // Kategori içindeki mevcut tercihleri koru
          [key]: value, // Belirli tercihi güncelle
        },
      };
    });
  }

  // preferences.studio.rosterChangeTypes'ın her zaman bir dizi olarak doğru şekilde başlatıldığından emin olun.
  // Bu, varsayılan değerin tam olarak reaktif olmaması veya dinamik olarak yeni onay kutuları eklenmesi durumunda gerekli olabilir.
  $: {
    if (preferences && !Array.isArray(preferences.studio.rosterChangeTypes)) {
      preferences.studio.rosterChangeTypes = [];
    }
  }

  // Roster Değişiklik Tipleri onay kutuları için değişiklikleri işleyen fonksiyon.
  // Seçilen değerleri diziye ekler veya çıkarır.
  function handleRosterChangeType(event: Event) {
    const target = event.target as HTMLInputElement;
    const value = target.value;
    if (target.checked) {
      if (!preferences.studio.rosterChangeTypes.includes(value)) {
        updatePreference("studio", "rosterChangeTypes", [
          ...preferences.studio.rosterChangeTypes,
          value,
        ]);
      }
    } else {
      updatePreference(
        "studio",
        "rosterChangeTypes",
        preferences.studio.rosterChangeTypes.filter((type) => type !== value)
      );
    }
  }

  function toggleTimezoneAirportSelection(event: Event) {
    const target = event.target as HTMLInputElement;

    if (target.value === "airport") {
      /* 1️⃣  Tipi “airport” yap */
      updatePreference("timezone", "type", "airport");

      /* 2️⃣  Daha önce bir havalimanı seçilmemiş-se ⇒ IST olarak başlat */
      if (!preferences.timezone.selectedAirport) {
        updatePreference("timezone", "selectedAirport", "IST");
      }
    } else {
      updatePreference("timezone", "type", "utc");
    }
  }

  // OK düğmesine basıldığında çalışacak fonksiyon.
  // Tercihler otomatik olarak mağazaya kaydedildiği için burada sadece pencereyi kapatma isteği gönderilir.
  function handleOk(): void {
    console.log(
      "Preferences saved (via store subscription). Dispatching closeoptions event."
    );
    showFeedbackMessage("Ayarlar başarıyla kaydedildi!", "success");
    // 'closeoptions' özel olayını üst bileşene gönderiyoruz.
    dispatch("closeoptions");
  }

  // Cancel düğmesine basıldığında çalışacak fonksiyon.
  // Bu durumda herhangi bir kaydetme yapılmaz, sadece pencereyi kapatma isteği gönderilir.
  function handleCancel(): void {
    console.log("Preferences cancelled. Dispatching closeoptions event.");
    showFeedbackMessage("Ayarlar iptal edildi.", "error");
    // 'closeoptions' özel olayını üst bileşene gönderiyoruz.
    dispatch("closeoptions");
  }

  // Kullanıcıya kısa süreli geri bildirim mesajı gösterme fonksiyonu
  function showFeedbackMessage(
    message: string,
    type: "success" | "error"
  ): void {
    feedbackMessage = message;
    feedbackType = type;
    showFeedback = true;
    setTimeout(() => (showFeedback = false), 3000); // 3 saniye sonra mesajı gizle
  }
</script>

<!-- user-options-window ana div'i, üst Window bileşeni tarafından kontrol edildiği için burada showUserOptionsWindow koşulu kaldırılmıştır. -->
<div
  class="user-options-window"
  style="background-color: var(--corporate-light-gray);"
>
  <div class="content-wrapper">
    <aside class="menu-bar">
      <ul>
        <li
          on:click={() => selectMenuItem("general")}
          class:active={activeMenuItem === "general"}
        >
          General
        </li>
        <li
          on:click={() => selectMenuItem("language")}
          class:active={activeMenuItem === "language"}
        >
          Language
        </li>
        <li
          on:click={() => selectMenuItem("opsview")}
          class:active={activeMenuItem === "opsview"}
        >
          Ops View
        </li>
        <li
          on:click={() => selectMenuItem("reporting")}
          class:active={activeMenuItem === "reporting"}
        >
          Reporting
        </li>
        <li
          on:click={() => selectMenuItem("studio")}
          class:active={activeMenuItem === "studio"}
        >
          Studio (Roster View)
        </li>
        <li
          on:click={() => selectMenuItem("timezone")}
          class:active={activeMenuItem === "timezone"}
        >
          TimeZone
        </li>
      </ul>
    </aside>

    <main class="main-content">
      {#if activeMenuItem === "general"}
        <div class="content-section">
          <h4>General Preferences</h4>

          <div class="preference-group">
            <!-- svelte-ignore a11y_label_has_associated_control -->
            <label>Windows Arrangement:</label>
            <label>
              <input
                type="radio"
                name="windowArrangement"
                value="overlapping"
                checked={preferences.general.windowArrangement ===
                  "overlapping"}
                on:change={(e) =>
                  updatePreference(
                    "general",
                    "windowArrangement",
                    e.currentTarget.value as "overlapping" | "sidebyside"
                  )}
              />
              Overlapping
            </label>
            <label>
              <input
                type="radio"
                name="windowArrangement"
                value="sidebyside"
                checked={preferences.general.windowArrangement === "sidebyside"}
                on:change={(e) =>
                  updatePreference(
                    "general",
                    "windowArrangement",
                    e.currentTarget.value as "overlapping" | "sidebyside"
                  )}
              />
              Side by side
            </label>
          </div>

          <div class="preference-group">
            <label>Gantt Scrolling and Zooming:</label>
            <label>
              <input
                type="radio"
                name="ganttScroll"
                value="scrollZoom"
                checked={preferences.general.ganttScroll === "scrollZoom"}
                on:change={(e) =>
                  updatePreference(
                    "general",
                    "ganttScroll",
                    e.currentTarget.value as "scrollZoom" | "zoomScroll"
                  )}
              />
              Scroll, ctrl to Zoom
            </label>
            <label>
              <input
                type="radio"
                name="ganttScroll"
                value="zoomScroll"
                checked={preferences.general.ganttScroll === "zoomScroll"}
                on:change={(e) =>
                  updatePreference(
                    "general",
                    "ganttScroll",
                    e.currentTarget.value as "scrollZoom" | "zoomScroll"
                  )}
              />
              Zoom, ctrl to Scroll
            </label>
          </div>

          <div class="preference-group">
            <label>
              <input
                type="checkbox"
                checked={preferences.general.showGC}
                on:change={(e) =>
                  updatePreference(
                    "general",
                    "showGC",
                    e.currentTarget.checked
                  )}
              />
              Show garbage collection tool
            </label>
          </div>
        </div>
      {:else if activeMenuItem === "language"}
        <div class="content-section">
          <h4>Language Preferences</h4>

          <div class="preference-group">
            <label for="locale">Locale:</label>
            <select
              id="locale"
              bind:value={preferences.language.locale}
              on:change={(e) =>
                updatePreference("language", "locale", e.currentTarget.value)}
            >
              <option value="default">none - use system default</option>
              <option value="en-US">en-US</option>
              <option value="tr-TR">tr-TR</option>
            </select>
          </div>

          <div class="preference-group">
            <label>
              <input
                type="checkbox"
                checked={preferences.language.showDefaultLabels}
                on:change={(e) =>
                  updatePreference(
                    "language",
                    "showDefaultLabels",
                    e.currentTarget.checked
                  )}
              />
              Display Default Labels in parentheses
            </label>
          </div>

          <div class="preference-group">
            <label>
              <input
                type="checkbox"
                checked={preferences.language.overrideCustomerLabels}
                on:change={(e) =>
                  updatePreference(
                    "language",
                    "overrideCustomerLabels",
                    e.currentTarget.checked
                  )}
              />
              Override Labels with Customer Specific
            </label>
          </div>
        </div>
      {:else if activeMenuItem === "opsview"}
        <div class="content-section">
          <h4>Ops View Preferences</h4>

          <div class="preference-group">
            <label>Start with Template:</label>
            <label>
              <input
                type="radio"
                name="opsTemplate"
                value="lastUsed"
                checked={preferences.opsview.startTemplate === "lastUsed"}
                on:change={(e) =>
                  updatePreference(
                    "opsview",
                    "startTemplate",
                    e.currentTarget.value as "lastUsed" | "startWith"
                  )}
              />
              Stored Last Used Template
            </label>
            <label>
              <input
                type="radio"
                name="opsTemplate"
                value="startWith"
                checked={preferences.opsview.startTemplate === "startWith"}
                on:change={(e) =>
                  updatePreference(
                    "opsview",
                    "startTemplate",
                    e.currentTarget.value as "lastUsed" | "startWith"
                  )}
              />
              Start with Template
              <select
                bind:value={preferences.opsview.templateName}
                on:change={(e) =>
                  updatePreference(
                    "opsview",
                    "templateName",
                    e.currentTarget.value
                  )}
              >
                <option>ops tiny</option>
                <!-- Other ops templates can be added here -->
              </select>
            </label>
          </div>
        </div>
      {:else if activeMenuItem === "reporting"}
        <div class="content-section">
          <h4>Reporting Preferences</h4>

          <div class="preference-group">
            <label for="urlParam">URL Parameter:</label>
            <input
              type="text"
              id="urlParam"
              bind:value={preferences.reporting.urlParam}
              on:change={(e) =>
                updatePreference(
                  "reporting",
                  "urlParam",
                  e.currentTarget.value
                )}
            />
          </div>

          <div class="preference-group">
            <label for="reportPath">Report Path:</label>
            <input
              type="text"
              id="reportPath"
              bind:value={preferences.reporting.reportPath}
              on:change={(e) =>
                updatePreference(
                  "reporting",
                  "reportPath",
                  e.currentTarget.value
                )}
            />
          </div>

          <div class="preference-group">
            <label for="rosterReportPath">Roster Reports Path:</label>
            <input
              type="text"
              id="rosterReportPath"
              bind:value={preferences.reporting.rosterReportPath}
              on:change={(e) =>
                updatePreference(
                  "reporting",
                  "rosterReportPath",
                  e.currentTarget.value
                )}
            />
          </div>

          <div class="preference-group">
            <label for="username">Username:</label>
            <input
              type="text"
              id="username"
              bind:value={preferences.reporting.username}
              on:change={(e) =>
                updatePreference(
                  "reporting",
                  "username",
                  e.currentTarget.value
                )}
            />
          </div>

          <div class="preference-group">
            <label for="password">Password:</label>
            <input
              type="password"
              id="password"
              bind:value={preferences.reporting.password}
              on:change={(e) =>
                updatePreference(
                  "reporting",
                  "password",
                  e.currentTarget.value
                )}
            />
          </div>
        </div>
      {:else if activeMenuItem === "studio"}
        <div class="content-section">
          <h4>Studio (Roster View) Preferences</h4>

          <div class="preference-group">
            <label>Start with Template:</label>
            <label>
              <input
                type="radio"
                name="templateStart"
                value="lastUsed"
                checked={preferences.studio.startTemplate === "lastUsed"}
                on:change={(e) =>
                  updatePreference(
                    "studio",
                    "startTemplate",
                    e.currentTarget.value as "lastUsed" | "startWith"
                  )}
              />
              Stored Last Used Template
            </label>
            <label>
              <input
                type="radio"
                name="templateStart"
                value="startWith"
                checked={preferences.studio.startTemplate === "startWith"}
                on:change={(e) =>
                  updatePreference(
                    "studio",
                    "startTemplate",
                    e.currentTarget.value as "lastUsed" | "startWith"
                  )}
              />
              Start with Template
              <select
                bind:value={preferences.studio.templateName}
                on:change={(e) =>
                  updatePreference(
                    "studio",
                    "templateName",
                    e.currentTarget.value
                  )}
              >
                <option>roster tiny</option>
                <!-- Other roster templates can be added here -->
              </select>
            </label>
          </div>

          <div class="preference-group">
            <label>
              <input
                type="checkbox"
                checked={preferences.studio.rememberZoom}
                on:change={(e) =>
                  updatePreference(
                    "studio",
                    "rememberZoom",
                    e.currentTarget.checked
                  )}
              />
              Remember last Zoom Factor
            </label>
          </div>

          <div class="preference-group slider-group">
            <label for="zoomFactor">Zoom Factor:</label>
            <input
              type="range"
              id="zoomFactor"
              min="0"
              max="10"
              step="0.1"
              bind:value={preferences.studio.zoomFactor}
              on:change={(e) =>
                updatePreference(
                  "studio",
                  "zoomFactor",
                  parseFloat(e.currentTarget.value)
                )}
            />
            <span>{preferences.studio.zoomFactor} days</span>
          </div>

          <div class="preference-group">
            <label for="defaultRoster">Default roster details level:</label>
            <select
              id="defaultRoster"
              bind:value={preferences.studio.defaultRosterLevel}
              on:change={(e) =>
                updatePreference(
                  "studio",
                  "defaultRosterLevel",
                  e.currentTarget.value
                )}
            >
              <option>PAIRING</option>
            </select>
          </div>

          <div class="preference-group">
            <label for="opsRoster">Ops to roster details level:</label>
            <select
              id="opsRoster"
              bind:value={preferences.studio.opsRosterLevel}
              on:change={(e) =>
                updatePreference(
                  "studio",
                  "opsRosterLevel",
                  e.currentTarget.value
                )}
            >
              <option>FLIGHT</option>
            </select>
          </div>

          <div class="preference-group">
            <label for="saturationLevel">
              Roster history saturation level for non changed:
            </label>
            <input
              type="number"
              id="saturationLevel"
              step="0.1"
              bind:value={preferences.studio.saturationLevel}
              on:change={(e) =>
                updatePreference(
                  "studio",
                  "saturationLevel",
                  parseFloat(e.currentTarget.value)
                )}
            />
          </div>

          <div class="preference-group types-of-change">
            <label>Types of Change for Roster History Light:</label>
            <div class="checkbox-list">
              <label>
                <input
                  type="checkbox"
                  value="assignments"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "assignments"
                  )}
                  on:change={handleRosterChangeType}
                />
                Assignments / Deassignments (included pairing edits)
              </label>
              <label>
                <input
                  type="checkbox"
                  value="scheduleChanges"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "scheduleChanges"
                  )}
                  on:change={handleRosterChangeType}
                />
                Schedule Changes / Movement Messages
              </label>
              <label>
                <input
                  type="checkbox"
                  value="notificationAcknowledgment"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "notificationAcknowledgment"
                  )}
                  on:change={handleRosterChangeType}
                />
                Notification Acknowledgment
              </label>
              <label>
                <input
                  type="checkbox"
                  value="checkInOutShowTime"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "checkInOutShowTime"
                  )}
                  on:change={handleRosterChangeType}
                />
                Check in / Check out / Show time
              </label>
              <label>
                <input
                  type="checkbox"
                  value="nonFlyingAssignments"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "nonFlyingAssignments"
                  )}
                  on:change={handleRosterChangeType}
                />
                Non-flying assignment changes
              </label>
              <label>
                <input
                  type="checkbox"
                  value="globalDelete"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "globalDelete"
                  )}
                  on:change={handleRosterChangeType}
                />
                Global Delete
              </label>
              <label>
                <input
                  type="checkbox"
                  value="rest"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "rest"
                  )}
                  on:change={handleRosterChangeType}
                />
                Rest
              </label>
              <label>
                <input
                  type="checkbox"
                  value="fillEmptyDays"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "fillEmptyDays"
                  )}
                  on:change={handleRosterChangeType}
                />
                Fill Empty Days
              </label>
              <label>
                <input
                  type="checkbox"
                  value="additionalNotification"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "additionalNotification"
                  )}
                  on:change={handleRosterChangeType}
                />
                Additional Notification
              </label>
              <label>
                <input
                  type="checkbox"
                  value="modifiedLegPosition"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "modifiedLegPosition"
                  )}
                  on:change={handleRosterChangeType}
                />
                Modified Leg Position
              </label>
              <label>
                <input
                  type="checkbox"
                  value="baseline"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "baseline"
                  )}
                  on:change={handleRosterChangeType}
                />
                Baseline
              </label>
              <label>
                <input
                  type="checkbox"
                  value="frms"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "frms"
                  )}
                  on:change={handleRosterChangeType}
                />
                FRMS
              </label>
              <label>
                <input
                  type="checkbox"
                  value="rejected"
                  checked={preferences.studio.rosterChangeTypes.includes(
                    "rejected"
                  )}
                  on:change={handleRosterChangeType}
                />
                Rejected
              </label>
            </div>
          </div>

          <div class="preference-group">
            <label>
              <input
                type="checkbox"
                checked={preferences.studio.showPatternTooltip}
                on:change={(e) =>
                  updatePreference(
                    "studio",
                    "showPatternTooltip",
                    e.currentTarget.checked
                  )}
              />
              Display pattern label tooltip on pucks
            </label>
          </div>

          <div class="preference-group">
            <label>
              <input
                type="checkbox"
                checked={preferences.studio.showNominationTooltip}
                on:change={(e) =>
                  updatePreference(
                    "studio",
                    "showNominationTooltip",
                    e.currentTarget.checked
                  )}
              />
              Display nomination tooltip on pucks
            </label>
          </div>

          <div class="preference-group">
            <label>
              <input
                type="checkbox"
                checked={preferences.studio.showBaselineTooltip}
                on:change={(e) =>
                  updatePreference(
                    "studio",
                    "showBaselineTooltip",
                    e.currentTarget.checked
                  )}
              />
              Display baseline tooltip on pucks
            </label>
          </div>
        </div>
      {:else if activeMenuItem === "timezone"}
        <div class="content-section">
          <h4>Time Zone Preferences</h4>

          <div class="preference-group">
            <label>Global Time Zone:</label>

            <!-- UTC seçeneği -->
            <label>
              <input
                type="radio"
                name="timezone"
                value="utc"
                checked={preferences.timezone.type === "utc"}
                on:change={toggleTimezoneAirportSelection}
              />
              UTC
            </label>

            <!-- Airport seçeneği -->
            <label>
              <input
                type="radio"
                name="timezone"
                value="airport"
                checked={preferences.timezone.type === "airport"}
                on:change={toggleTimezoneAirportSelection}
              />
              Airport
              <select
                disabled={preferences.timezone.type !== "airport"}
                bind:value={preferences.timezone.selectedAirport}
                on:change={(e) =>
                  updatePreference(
                    "timezone",
                    "selectedAirport",
                    e.currentTarget.value
                  )}
              >
                <option value="">(Select Airport)</option>
                <option value="IST">Istanbul - IST (GMT+3)</option>
                <option value="JFK">New York - JFK (GMT-4)</option>
                <option value="LHR">London - LHR (GMT+0)</option>
                <option value="DXB">Dubai - DXB (GMT+4)</option>
              </select>
            </label>
          </div>
        </div>
      {/if}
    </main>
  </div>

  <footer class="footer-buttons">
    <button class="ok-button" on:click={handleOk}>OK</button>
    <button class="cancel-button" on:click={handleCancel}>Cancel</button>
  </footer>

  <!-- Geri bildirim mesajı, sadece showFeedback true olduğunda görünür -->
  {#if showFeedback}
    <div class="feedback-message {feedbackType}">
      {feedbackMessage}
    </div>
  {/if}
</div>

<style>
  /* Tercih grubu elemanları arasındaki boşluk */
  .preference-group {
    margin-bottom: 20px;
  }

  /* Tercih grubu etiketleri için hizalama ve stil */
  .preference-group label {
    display: inline-flex;
    align-items: center;
    margin-right: 20px;
    font-weight: normal;
    color: #333;
  }

  /* Ana kullanıcı seçenekleri penceresi stil ayarları */
  .user-options-window {
    display: flex;
    flex-direction: column; /* İçeriği dikey olarak hizala */
    height: 100%; /* Mevcut yüksekliğin tamamını kapla */
    /* Bu stiller artık Window.svelte tarafından yönetilecek, burada kaldırıldı veya pasif bırakıldı */
    /* border: 1px solid #ccc; */
    /* box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1); */
    /* background-color: #fff; */
    overflow: hidden; /* İçerik taşmasını önler */
    border-radius: 4px; /* Köşeleri yuvarlar (isteğe bağlı, Window tarafından da kontrol edilebilir) */
    font-family: "Inter", sans-serif;
    color: var(--corporate-text-color);
  }

  /* İçerik sarıcı stil ayarları (menü çubuğu ve ana içerik için) */
  .content-wrapper {
    display: flex;
    flex: 1; /* Kalan dikey alanı doldurmasını sağlar */
    overflow: hidden; /* Kaydırılabilir alanları içermesi için önemli */
  }

  /* Menü çubuğu stil ayarları */
  .menu-bar {
    width: 200px; /* Sabit genişlik */
    background-color: var(--corporate-header-bg); /* Açık gri arka plan */
    border-right: 1px solid var(--corporate-medium-gray); /* Sağda ince kenarlık */
    overflow-y: auto; /* Dikey olarak kaydırılabilir hale getirir */
    padding: 15px 0; /* Üst ve alt iç boşluk */
  }

  /* Menü çubuğu listesi stil ayarları */
  .menu-bar ul {
    list-style: none; /* Madde işaretlerini kaldırır */
    padding: 0; /* İç boşluğu kaldırır */
    margin: 0; /* Kenar boşluğunu kaldırır */
  }

  /* Menü çubuğu listesi öğeleri stil ayarları */
  .menu-bar li {
    padding: 12px 20px; /* İç boşluk */
    cursor: pointer; /* Fare imlecini işaretçi yapar */
    border-bottom: 1px solid var(--corporate-medium-gray); /* Alt kenarlık */
    transition: background-color var(--corporate-transition-speed)
      var(--corporate-transition-timing); /* Arka plan rengi değişimi için animasyon */
  }

  /* Menü çubuğu öğesinin üzerine gelindiğinde arka plan rengini değiştir */
  .menu-bar li:hover {
    background-color: var(
      --corporate-medium-gray
    ); /* Daha koyu gri arka plan */
  }

  /* Aktif menü öğesi stil ayarları */
  .menu-bar li.active {
    background-color: var(--corporate-light-gray); /* Farklı arka plan rengi */
    font-weight: bold; /* Kalın yazı tipi */
    color: var(--corporate-accent-blue); /* Mavi renk (aktif öğeyi vurgular) */
  }

  /* Ana içerik alanı stil ayarları */
  .main-content {
    flex: 1; /* Kalan yatay alanı doldurmasını sağlar */
    overflow-y: auto; /* Dikey olarak kaydırılabilir hale getirir */
    padding: 20px; /* İç boşluk */
  }

  /* İçerik bölümü stil ayarları */
  .content-section {
    margin-bottom: 20px; /* Alt kenar boşluğu */
  }

  /* İçerik bölümü başlığı stil ayarları */
  .content-section h4 {
    margin-top: 0; /* Üst kenar boşluğunu kaldırır */
    color: var(--corporate-text-color);
    margin-bottom: 15px; /* Alt kenar boşluğu */
  }

  /* İçerik bölümü etiketleri stil ayarları */
  .content-section label {
    display: block; /* Bloğu kaplar */
    margin-bottom: 5px; /* Alt kenar boşluğu */
    font-weight: bold; /* Kalın yazı tipi */
    color: var(--corporate-text-color);
  }

  /* Metin ve şifre giriş alanları stil ayarları */
  .content-section input[type="text"],
  .content-section input[type="password"],
  .content-section input[type="number"],
  .content-section select {
    width: calc(
      100% - 20px
    ); /* Kalan genişliği kaplar (padding hesaba katılır) */
    padding: 10px; /* İç boşluk */
    margin-bottom: 15px; /* Alt kenar boşluğu */
    border: 1px solid var(--corporate-medium-gray); /* İnce gri kenarlık */
    border-radius: 4px; /* Köşeleri yuvarlar */
    background-color: white;
    color: var(--corporate-text-color);
  }

  .content-section input[type="range"] {
    width: calc(100% - 20px);
    margin-bottom: 15px;
  }

  /* Alt kısım düğmeleri alanı stil ayarları */
  .footer-buttons {
    padding: 15px 20px; /* İç boşluk */
    border-top: 1px solid var(--corporate-medium-gray); /* Üst kenarlık */
    background-color: var(--corporate-header-bg); /* Açık gri arka plan */
    text-align: right; /* Metni sağa hizalar */
    flex-shrink: 0;
  }

  /* Alt kısım düğmeleri stil ayarları */
  .footer-buttons button {
    padding: 10px 20px; /* İç boşluk */
    margin-left: 10px; /* Soldan kenar boşluğu */
    border: none; /* Kenarlığı kaldırır */
    border-radius: 4px; /* Köşeleri yuvarlar */
    cursor: pointer; /* Fare imlecini işaretçi yapar */
    font-size: 1rem; /* Yazı tipi boyutu */
    transition: background-color var(--corporate-transition-speed)
      var(--corporate-transition-timing); /* Arka plan rengi değişimi için animasyon */
  }

  /* "OK" düğmesi stil ayarları */
  .ok-button {
    background-color: var(--corporate-accent-blue); /* Mavi arka plan */
    color: white; /* Beyaz yazı tipi */
  }

  /* "OK" düğmesinin üzerine gelindiğinde arka plan rengini değiştir */
  .ok-button:hover {
    background-color: #0056b3; /* Daha koyu mavi */
  }

  /* "Cancel" düğmesi stil ayarları */
  .cancel-button {
    background-color: var(--corporate-dark-gray); /* Koyu gri arka plan */
    color: white; /* Beyaz yazı tipi */
  }

  /* "Cancel" düğmesinin üzerine gelindiğinde arka plan rengini değiştir */
  .cancel-button:hover {
    background-color: #5a6268; /* Daha koyu gri */
  }

  /* Geri bildirim mesajı stili */
  .feedback-message {
    position: absolute;
    bottom: 80px; /* Düğmelerin üstünde ve alt çubuğun üstünde */
    left: 50%;
    transform: translateX(-50%);
    padding: 10px 20px;
    border-radius: 5px;
    font-weight: bold;
    color: white;
    z-index: 1000;
    opacity: 1;
    transition: opacity 0.5s ease-out;
    white-space: nowrap; /* Mesajın tek satırda kalmasını sağlar */
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
  }

  .feedback-message.success {
    background-color: #28a745; /* Yeşil */
  }

  .feedback-message.error {
    background-color: #dc3545; /* Kırmızı */
  }
</style>
