<script lang="ts">
  import { setAuth } from "@/stores/authStore"; // setAuth'ı import edin
  import { setStatus } from "@/stores/appStates";
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();

  let username = "";
  let password = "";
  let errorMessage = "";
  let isLoading = false;
  let showPassword = false;

  function toggleShowPassword() {
    showPassword = !showPassword;
  }

  async function handleLogin() {
    errorMessage = "";
    isLoading = true;

    // İstemci tarafı basit doğrulama
    if (!username.trim() || !password.trim()) {
      errorMessage = "Kullanıcı adı ve şifre boş bırakılamaz.";
      setStatus("Giriş başarısız: Boş alanlar var.");
      isLoading = false;
      return;
    }

    try {
      // API_BASE_URL'in doğru tanımlandığından emin olun (genellikle bir env değişkeni)
      const API_BASE_URL = import.meta.env.VITE_API_URL;
      const response = await fetch(`${API_BASE_URL}/api/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, password }),
      });

      if (response.ok) {
        const data = await response.json();
        const token = data.token;
        const userId = data.user_id;
        const username = data.username;

        // setAuth çağrılıyor. Bu fonksiyonun kendisi artık token'ı localStorage'a kaydetmiyor.
        setAuth(userId, username, token);

        setStatus("Giriş başarılı!");
        dispatch("loginSuccess"); // Ana bileşene girişin başarılı olduğunu bildir
      } else {
        const errorData = await response.json();
        errorMessage = errorData.error || "Kullanıcı adı veya şifre yanlış.";
        setStatus("Giriş başarısız.");
      }
    } catch (error) {
      console.error("Giriş sırasında ağ hatası oluştu:", error);
      errorMessage =
        "Sunucuya bağlanılamadı. Lütfen internet bağlantınızı kontrol edin.";
      setStatus("Giriş sırasında hata oluştu.");
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="login-container">
  <div class="login-box" role="dialog" aria-labelledby="login-title">
    <h2 id="login-title" class="login-title">Giriş Yap</h2>
    <form on:submit|preventDefault={handleLogin}>
      <div class="input-group">
        <label for="username">Kullanıcı Adı:</label>
        <input
          type="text"
          id="username"
          bind:value={username}
          required
          aria-required="true"
          aria-label="Kullanıcı Adı"
          disabled={isLoading}
        />
      </div>
      <div class="input-group password-wrapper">
        <label for="password">Şifre:</label>
        <input
          type={showPassword ? "text" : "password"}
          id="password"
          bind:value={password}
          required
          aria-required="true"
          aria-label="Şifre"
          aria-describedby="password-visibility-help"
          disabled={isLoading}
          class="has-toggle"
        />
        <button
          type="button"
          class="toggle-password-btn"
          on:click={toggleShowPassword}
          aria-label={showPassword ? "Şifreyi gizle" : "Şifreyi göster"}
          aria-pressed={showPassword}
          title={showPassword ? "Şifreyi gizle" : "Şifreyi göster"}
        >
          {#if showPassword}
            <i class="fas fa-eye-slash" aria-hidden="true"></i>
          {:else}
            <i class="fas fa-eye" aria-hidden="true"></i>
          {/if}
        </button>
        <span id="password-visibility-help" class="sr-only">
          Bu düğme şifrenin görünürlüğünü değiştirir.
        </span>
      </div>

      {#if errorMessage}
        <p class="error-message" role="alert" aria-live="assertive">
          {errorMessage}
        </p>
      {/if}
      <button
        type="submit"
        class="login-button"
        disabled={isLoading}
        aria-live="polite"
      >
        {#if isLoading}
          <i class="fas fa-spinner fa-spin login-spinner" aria-hidden="true"
          ></i>
          Giriş Yapılıyor...
        {:else}
          Giriş
        {/if}
      </button>
    </form>
  </div>
</div>

<style>
  /* Font Awesome için import eklediğinizden emin olun (global CSS'inizde yoksa) */
  @import url("https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.4/css/all.min.css");

  .login-container {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background-color: var(--thy-primary-red);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 9999;
    transition: background-color 0.3s ease;
  }

  .login-box {
    background-color: var(--thy-white);
    padding: 2.5rem;
    border-radius: 8px;
    box-shadow: var(--thy-shadow-strong);
    text-align: center;
    width: 100%;
    max-width: 400px;
    opacity: 0;
    transform: translateY(-20px);
    animation: fadeIn 0.5s forwards ease-out;
  }

  @keyframes fadeIn {
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .login-title {
    color: var(--thy-primary-red);
    margin-bottom: 2rem;
    font-size: 1.8rem;
    font-weight: 700;
  }

  .input-group {
    margin-bottom: 1.2rem;
    text-align: left;
  }

  .input-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: var(--thy-text-color);
  }

  .input-group input {
    width: 100%;
    padding: 0.8rem 1rem;
    border: 1px solid var(--thy-border-color);
    border-radius: 4px;
    font-size: 1rem;
    box-sizing: border-box;
    transition: border-color 0.2s ease-in-out;
  }

  .input-group input:focus {
    outline: none;
    border-color: var(--thy-secondary-blue);
    box-shadow: 0 0 0 2px rgba(0, 75, 133, 0.2);
  }

  .error-message {
    color: var(--thy-primary-red);
    background-color: #ffe0e0;
    border: 1px solid var(--thy-primary-red);
    padding: 0.75rem 1rem;
    border-radius: 4px;
    margin-top: -0.8rem;
    margin-bottom: 1rem;
    font-size: 0.9rem;
    text-align: left;
    animation: shake 0.3s ease-in-out;
  }

  @keyframes shake {
    0%,
    100% {
      transform: translateX(0);
    }
    20%,
    60% {
      transform: translateX(-5px);
    }
    40%,
    80% {
      transform: translateX(5px);
    }
  }

  .login-button {
    background-color: var(--thy-secondary-blue);
    color: var(--thy-white);
    border: none;
    padding: 0.8rem 2rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 1.1rem;
    font-weight: 600;
    transition:
      background-color 0.2s ease-in-out,
      box-shadow 0.2s ease-in-out;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .login-button:hover:not(:disabled) {
    background-color: #003a66;
    box-shadow: var(--thy-shadow-medium);
  }

  .login-button:active:not(:disabled) {
    background-color: #002d4d;
  }

  .login-button:disabled {
    background-color: var(--thy-medium-gray);
    cursor: not-allowed;
    opacity: 0.8;
  }

  .login-spinner {
    margin-right: 0.5rem;
  }
  .password-wrapper {
    position: relative;
  }

  .input-group input.has-toggle {
    padding-right: 2.5rem; /* butona yer aç */
  }

  .toggle-password-btn {
    position: absolute;
    right: 0.6rem;
    top: 70%;
    transform: translateY(-50%);
    border: none;
    background: transparent;
    cursor: pointer;
    padding: 0.25rem;
    line-height: 1;
  }

  .toggle-password-btn:focus {
    outline: none;
    box-shadow: 0 0 0 2px rgba(0, 75, 133, 0.2);
    border-radius: 4px;
  }

  /* Erişilebilirlik için sadece-ekran-okuyucu sınıfı */
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 1px, 1px);
    white-space: nowrap;
    border: 0;
  }
</style>
