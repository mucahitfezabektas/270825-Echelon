from zeep import Client, Settings
from zeep.plugins import HistoryPlugin
import hashlib
import base64
import time
from datetime import datetime, timedelta

# === TEST ORTAMI BİLGİLERİ ===
CLIENT_CODE = "10738"
CLIENT_USERNAME = "Test"
CLIENT_PASSWORD = "Test"
GUID = "0c13d406-873b-403b-9c09-a5766840d98c"
WSDL_URL = "https://testposws.param.com.tr/turkpos.ws/service_turkpos_prod.asmx?wsdl"

# === GÜVENLİK NESNESİ ===
G = {
    "CLIENT_CODE": CLIENT_CODE,
    "CLIENT_USERNAME": CLIENT_USERNAME,
    "CLIENT_PASSWORD": CLIENT_PASSWORD,
}

# === SOAP CLIENT HAZIRLIĞI ===
history = HistoryPlugin()
settings = Settings(strict=False, xml_huge_tree=True)
client = Client(wsdl=WSDL_URL, settings=settings, plugins=[history])

# Ortak test kart bilgileri
TEST_CARD_NO = "4546711234567894"
TEST_CARD_EXP_MONTH = "12"
TEST_CARD_EXP_YEAR = "2026"
TEST_CARD_CVC = "000"  # Dokümanda CVC, Zeep hata mesajında CVV olarak geçebilir.
TEST_CARD_OWNER_GSM = "5555555555"
TEST_CARD_OWNER_NAME = "TEST KART"

# Ortak URL'ler
Hata_URL = "https://example.com/hata"
Basarili_URL = "https://example.com/basarili"
Ref_URL = "https://example.com"
IPAdr = "1.2.3.4"


# === HASH HESAPLAMA FONKSİYONU ===
# Dokümanda belirtilen SHA2B64 metoduna benzer şekilde hash oluşturur.
def calculate_hash(params_str):
    sha1_bytes = hashlib.sha1(params_str.encode("utf-8")).digest()
    return base64.b64encode(sha1_bytes).decode()


# === 1. POS ÖDEME TESTİ ===
print("🚀 1. POS ÖDEME TESTİ BAŞLIYOR...")

# Siparis_ID'yi her zaman benzersiz tutmak için zaman damgası kullanıyoruz
pos_odeme_siparis_id = f"POS-TEST-{int(time.time())}"
pos_odeme_taksit = "1"
pos_odeme_islem_tutar = "100,00"
pos_odeme_toplam_tutar = "100,00"

# POS Ödeme için hash hesaplama (Doküman sayfa 5)
raw_hash_pos_odeme = f"{CLIENT_CODE}{GUID}{pos_odeme_taksit}{pos_odeme_islem_tutar}{pos_odeme_toplam_tutar}{pos_odeme_siparis_id}{Hata_URL}{Basarili_URL}"
pos_odeme_islem_hash = calculate_hash(raw_hash_pos_odeme)

odeme_response = None
try:
    odeme_response = client.service.Pos_Odeme(
        G=G,
        GUID=GUID,
        KK_Sahibi=TEST_CARD_OWNER_NAME,
        KK_No=TEST_CARD_NO,
        KK_SK_Ay=TEST_CARD_EXP_MONTH,
        KK_SK_Yil=TEST_CARD_EXP_YEAR,
        KK_CVC=TEST_CARD_CVC,
        KK_Sahibi_GSM=TEST_CARD_OWNER_GSM,
        Hata_URL=Hata_URL,
        Basarili_URL=Basarili_URL,
        Siparis_ID=pos_odeme_siparis_id,
        Siparis_Aciklama="Test ödemesi",
        Taksit=int(pos_odeme_taksit),
        Islem_Tutar=pos_odeme_islem_tutar,
        Toplam_Tutar=pos_odeme_toplam_tutar,
        Islem_Hash=pos_odeme_islem_hash,
        Islem_Guvenlik_Tip="NS",
        Islem_ID="",
        IPAdr=IPAdr,
        Ref_URL=Ref_URL,
    )

    print("--- POS Ödeme Sonucu ---")
    print("Sonuç:", odeme_response.Sonuc)
    print("Açıklama:", odeme_response.Sonuc_Str)
    print("Islem_ID:", odeme_response.Islem_ID)

except Exception as e:
    print(f"⚠️ POS ÖDEME TESTİ SIRASINDA HATA OLUŞTU: {e}")

# === 2. KISMİ İADE TESTİ (Eğer ödeme başarılıysa) ===
if odeme_response and str(odeme_response.Sonuc) == "1":
    print("\n🔁 2. KISMİ İADE TESTİ BAŞLIYOR...")

    # İade edilecek tutar, dokümanda belirtildiği gibi Double (float) formatında olmalı (Doküman sayfa 17)
    iade_tutari = 50.00

    try:
        iade_response = client.service.TP_Islem_Iptal_Iade_Kismi2(
            G=G,
            GUID=GUID,
            Durum="IADE",  # İade işlemi için "IADE" değeri gönderilir
            Siparis_ID=pos_odeme_siparis_id,  # Ödeme yapılırken kullanılan Siparis_ID
            Tutar=iade_tutari,  # İade edilecek tutar (float olarak)
        )

        print("--- İPTAL/İADE Sonucu ---")
        print("Sonuç:", iade_response.Sonuc)
        print("Açıklama:", iade_response.Sonuc_Str)
    except Exception as e:
        print(f"⚠️ KISMİ İADE TESTİ SIRASINDA HATA OLUŞTU: {e}")
else:
    print("⚠️ POS Ödeme başarısız olduğu için iade testi atlandı.")


# === 3. KART SAKLAMA TESTİ (KK_Saklama) ===
print("\n💳 3. KART SAKLAMA TESTİ BAŞLIYOR (KK_Saklama)...")
# Doküman sayfa 30 ve Zeep hata mesajı dikkate alınarak parametreler güncellendi.
ks_guid = None
try:
    card_storage_response = client.service.KK_Saklama(
        G=G,
        # GUID parametresi dokümanda var ancak hata mesajındaki WSDL imzasında yok.
        # Zeep hatasına göre kaldırıldı.
        Kart_No=TEST_CARD_NO,  # Hata mesajındaki WSDL imzasında var.
        KK_Sahibi=TEST_CARD_OWNER_NAME,
        KK_No=TEST_CARD_NO,
        KK_SK_Ay=TEST_CARD_EXP_MONTH,
        KK_SK_Yil=TEST_CARD_EXP_YEAR,
        KK_CVV=TEST_CARD_CVC,  # Dokümanda KK_CVC, hata mesajında KK_CVV. Hata mesajına göre güncellendi.
        Data1="",  # Hata mesajındaki WSDL imzasında var.
        Data2="",  # Hata mesajındaki WSDL imzasında var.
        Data3="",  # Hata mesajındaki WSDL imzasında var.
        # KK_Kart_Adi ve KK_Islem_ID dokümanda opsiyonel olsa da, hata mesajındaki WSDL imzasında yok.
        # Bu nedenle kaldırıldı.
    )

    print("--- Kart Saklama Sonucu ---")
    print("Sonuç:", card_storage_response.Sonuc)
    print("Açıklama:", card_storage_response.Sonuc_Str)
    # Çıktıya göre KS_GUID yerine 'GUID' niteliği dönüyor.
    if str(card_storage_response.Sonuc) == "1" and hasattr(
        card_storage_response, "GUID"
    ):
        ks_guid = card_storage_response.GUID
        print(f"KS_GUID (Alınan): {ks_guid}")
    else:
        print("KS_GUID alınamadı.")

except Exception as e:
    print(f"⚠️ KART SAKLAMA TESTİ SIRASINDA HATA OLUŞTU: {e}")

# === 4. SAKLI KART İLE ÖDEME TESTİ (KS_Tahsilat) ===
if ks_guid:
    print("\n💰 4. SAKLI KART İLE ÖDEME TESTİ BAŞLIYOR (KS_Tahsilat)...")
    # Doküman sayfa 31'deki metot adı KS_Tahsilat. Ancak "Service has no operation" hatası alıyoruz.
    # TP_Islem_Iptal_Iade_Kismi2 örneğine benzer şekilde TP_ öneki ile deniyoruz.
    ks_tahsilat_siparis_id = f"KS-TEST-{int(time.time())}"
    ks_tahsilat_taksit = "1"
    ks_tahsilat_islem_tutar = "75,00"
    ks_tahsilat_toplam_tutar = "75,00"  # Komisyon dahil tutar

    # KS_Tahsilat için hash hesaplama (Pos_Odeme ile benzer formatı kullanıyoruz, doküman sayfa 5)
    raw_hash_ks_tahsilat = f"{CLIENT_CODE}{GUID}{ks_tahsilat_taksit}{ks_tahsilat_islem_tutar}{ks_tahsilat_toplam_tutar}{ks_tahsilat_siparis_id}{Hata_URL}{Basarili_URL}"
    ks_tahsilat_islem_hash = calculate_hash(raw_hash_ks_tahsilat)

    try:
        # Metot adını dokümandaki orijinal haline geri çeviriyoruz
        ks_tahsilat_response = client.service.KS_Tahsilat(
            G=G,
            GUID=GUID,
            KS_GUID=ks_guid,
            CW="",  # NonSecure işlem için boş bırakılabilir
            KK_Sahibi_GSM=TEST_CARD_OWNER_GSM,
            Hata_URL=Hata_URL,
            Basarili_URL=Basarili_URL,
            Siparis_ID=ks_tahsilat_siparis_id,
            Siparis_Aciklama="Saklı kart ile test ödemesi",
            Taksit=int(ks_tahsilat_taksit),
            Islem_Tutar=ks_tahsilat_islem_tutar,
            Toplam_Tutar=ks_tahsilat_toplam_tutar,
            Islem_Guvenlik_Tip="NS",
            Islem_ID="",
            IPAdr=IPAdr,
            Ref_URL=Ref_URL,
        )

        print("--- Saklı Kart ile Ödeme Sonucu ---")
        print("Sonuç:", ks_tahsilat_response.Sonuc)
        print("Açıklama:", ks_tahsilat_response.Sonuc_Str)
        print("Islem_ID:", ks_tahsilat_response.Islem_ID)
    except Exception as e:
        print(f"⚠️ SAKLI KART İLE ÖDEME TESTİ SIRASINDA HATA OLUŞTU: {e}")
else:
    print("⚠️ Kart saklama başarılı olmadığı için saklı kart ile ödeme testi atlandı.")


# === 5. SAKLI KARTLARI LİSTELEME TESTİ (KK_Sakli_Liste) ===
print("\n📋 5. SAKLI KARTLARI LİSTELEME TESTİ BAŞLIYOR (KK_Sakli_Liste)...")
# Doküman sayfa 34
try:
    card_list_response = client.service.KK_Sakli_Liste(
        G=G,
        Kart_No=TEST_CARD_NO,  # Test kart numarasını kullanarak listeleyelim
        KS_KK_Kisi_ID="",  # Opsiyonel
    )

    print("--- Saklı Kart Listesi Sonucu ---")
    print("Sonuç:", card_list_response.Sonuc)
    print("Açıklama:", card_list_response.Sonuc_Str)
    if (
        str(card_list_response.Sonuc) == "1"
        and hasattr(card_list_response, "DT_Bilgi")
        and card_list_response.DT_Bilgi
    ):
        print("Saklı Kartlar:")
        # DT_Bilgi nesnesinin çıktısı incelendiğinde, kart bilgilerinin 'Temp' anahtarı altında olduğu görüldü.
        # Ayrıca, liste '_value_1' altında ve her 'Temp' nesnesi de bir liste içinde.
        # Bu yapıyı doğru şekilde parse etmek için aşağıdaki döngüyü kullanıyoruz.
        parsed_cards = []
        if hasattr(card_list_response.DT_Bilgi, "_value_1") and isinstance(
            card_list_response.DT_Bilgi._value_1, list
        ):
            for item in card_list_response.DT_Bilgi._value_1:
                # Her bir item'ın içinde bir sözlük ve onun içinde de '_value_1' ve 'Temp' olabilir.
                if (
                    isinstance(item, dict)
                    and "_value_1" in item
                    and isinstance(item["_value_1"], list)
                ):
                    for sub_item in item["_value_1"]:
                        if isinstance(sub_item, dict) and "Temp" in sub_item:
                            parsed_cards.append(sub_item["Temp"])
                elif hasattr(item, "Temp"):  # Eğer doğrudan Temp objesi varsa
                    parsed_cards.append(item.Temp)

        if parsed_cards:
            for card in parsed_cards:
                # Kart niteliklerine doğru yoldan erişiyoruz
                print(
                    f"  ID: {card.ID}, GUID: {card.KK_GUID}, Tarih: {card.Tarih.strftime('%d.%m.%Y %H:%M:%S')}, Kart No: {card.KK_No}, Kart Adı: {card.Kart_Adi if hasattr(card, 'Kart_Adi') and card.Kart_Adi else 'Yok'}"
                )
                print(
                    f"    Banka: {card.KK_Banka}, Marka: {card.KK_Marka}, Tip: {card.KK_Tip}, Son 4 Hane: {card.KK_Son4}"
                )
        else:
            print("  Hiç saklı kart bulunamadı veya ayrıştırma başarısız oldu.")
    else:
        print("Saklı kart bulunamadı veya listeleme başarısız.")
except Exception as e:
    print(f"⚠️ SAKLI KARTLARI LİSTELEME TESTİ SIRASINDA HATA OLUŞTU: {e}")


# === 6. SAKLI KART SİLME TESTİ (KS_Kart_Sil) ===
if ks_guid:
    print("\n🗑️ 6. SAKLI KART SİLME TESTİ BAŞLIYOR (KS_Kart_Sil)...")
    # Doküman sayfa 38'deki metot adı KS_Kart_Sil. Ancak "Service has no operation" hatası alıyoruz.
    # TP_Islem_Iptal_Iade_Kismi2 örneğine benzer şekilde TP_ öneki ile deniyoruz.
    try:
        # Metot adını dokümandaki orijinal haline geri çeviriyoruz
        delete_card_response = client.service.KS_Kart_Sil(
            G=G,
            KS_GUID=ks_guid,  # Sakladığımız kartın GUID'sini kullanarak siliyoruz
            KK_Islem_ID="",  # Opsiyonel
        )

        print("--- Saklı Kart Silme Sonucu ---")
        print("Sonuç:", delete_card_response.Sonuc)
        print("Açıklama:", delete_card_response.Sonuc_Str)
    except Exception as e:
        print(f"⚠️ SAKLI KART SİLME TESTİ SIRASINDA HATA OLUŞTU: {e}")
else:
    print("⚠️ Silinecek kart GUID'si bulunamadığı için kart silme testi atlandı.")


# === 7. MUTABAKAT ÖZETİ TESTİ (TP_Mutabakat_Ozet) ===
print("\n📊 7. MUTABAKAT ÖZETİ TESTİ BAŞLIYOR (TP_Mutabakat_Ozet)...")
# Doküman sayfa 54
# Son 24 saatlik özeti alalım
end_date = datetime.now()
start_date = end_date - timedelta(days=1)

# Tarih formatı: dd.MM.yyyy HH:mm:ss
tarih_bas = start_date.strftime("%d.%m.%Y %H:%M:%S")
tarih_bit = end_date.strftime("%d.%m.%Y %H:%M:%S")

try:
    mutabakat_ozet_response = client.service.TP_Mutabakat_Ozet(
        G=G, GUID=GUID, Tarih_Bas=tarih_bas, Tarih_Bit=tarih_bit
    )

    print("--- Mutabakat Özeti Sonucu ---")
    print("Sonuç:", mutabakat_ozet_response.Sonuc)
    print("Açıklama:", mutabakat_ozet_response.Sonuc_Str)
    if (
        str(mutabakat_ozet_response.Sonuc) == "1"
        and hasattr(mutabakat_ozet_response, "DT_Bilgi")
        and mutabakat_ozet_response.DT_Bilgi
    ):
        print("Mutabakat Detayları (DT_Bilgi):")
        # DT_Bilgi'nin yapısını kontrol etmek gerekebilir, dokümanda "Saklanmış kart listesi" yazıyor ama
        # Mutabakat Özeti için işlem özetleri dönmesi beklenir.
        # Burada örnek olarak sadece varlığını kontrol ediyoruz.
        print(
            f"  DT_Bilgi nesnesi mevcut ve içeriği: {mutabakat_ozet_response.DT_Bilgi}"
        )
    else:
        print("Mutabakat özeti bulunamadı veya işlem başarısız.")
except Exception as e:
    print(f"⚠️ MUTABAKAT ÖZETİ TESTİ SIRASINDA HATA OLUŞTU: {e}")


# === SOAP TRACE ===
print("\n--- Gönderilen SOAP ---")
if history.last_sent and "envelope" in history.last_sent:
    print(
        str(history.last_sent["envelope"].decode("utf-8"))
        if hasattr(history.last_sent["envelope"], "decode")
        else str(history.last_sent["envelope"])
    )
else:
    print("Gönderilen SOAP isteği bulunamadı veya boş.")

print("\n--- Gelen SOAP ---")
if history.last_received and "envelope" in history.last_received:
    print(
        str(history.last_received["envelope"].decode("utf-8"))
        if hasattr(history.last_received["envelope"], "decode")
        else str(history.last_received["envelope"])
    )
else:
    print("Gelen SOAP yanıtı bulunamadı veya boş.")