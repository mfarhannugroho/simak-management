<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const localError = ref<string | null>(null)

async function handleLogin() {
  localError.value = null
  if (!email.value || !password.value) {
    localError.value = 'Email dan kata sandi wajib diisi'
    return
  }

  try {
    const success = await authStore.login({
      email: email.value.trim(),
      password: password.value,
    })
    if (success) {
      const redirect = (route.query.redirect as string) || '/'
      router.push(redirect)
    }
  } catch (err: any) {
    localError.value = err.message || 'Login gagal, periksa email dan kata sandi'
  }
}
</script>

<template>
  <div class="min-h-screen w-full bg-slate-900 flex flex-col justify-center items-center p-4 relative overflow-hidden">
    <!-- Background Architectural Accent Patterns -->
    <div class="absolute -top-32 -left-32 w-96 h-96 bg-inka-700/20 rounded-full blur-3xl pointer-events-none"></div>
    <div class="absolute -bottom-32 -right-32 w-96 h-96 bg-blue-600/10 rounded-full blur-3xl pointer-events-none"></div>

    <div class="w-full max-w-4xl relative z-10 grid grid-cols-1 lg:grid-cols-12 gap-8 items-center">
      <!-- Left: Brand Story & Demo Switcher (5 cols) -->
      <div class="lg:col-span-5 text-white space-y-6">
        <div class="flex items-center gap-3">
          <div class="h-12 px-3 py-1.5 rounded-xl bg-white flex items-center justify-center shadow-lg border border-white/20 shrink-0">
            <img src="/logo.png" alt="Logo Resmi PT INKA (Persero)" class="h-8 w-auto object-contain" />
          </div>
          <div>
            <h1 class="text-xl font-bold tracking-tight text-white leading-tight">
              PT INKA (Persero)
            </h1>
            <p class="text-xs font-semibold text-inka-400 uppercase tracking-widest">
              SIMAK Digital Hub
            </p>
          </div>
        </div>

        <div class="space-y-2">
          <h2 class="text-2xl font-extrabold text-white leading-snug">
            Sistem Perizinan & Verifikasi Barang Bawaan Kerja
          </h2>
          <p class="text-xs text-slate-300 leading-relaxed">
            Platform terpadu pencatatan alat kerja teknisi, alur persetujuan berjenjang, penerbitan dokumen QR resmi, dan verifikasi pos keamanan pabrik.
          </p>
        </div>

        <!-- Corporate Standards Highlights (Production) -->
        <div class="space-y-3 pt-2">
          <div class="flex items-start gap-3 p-3 rounded-xl bg-slate-800/60 border border-slate-700/60">
            <div class="w-8 h-8 rounded-lg bg-inka-900/60 text-inka-400 flex items-center justify-center shrink-0 border border-inka-700/40">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
              </svg>
            </div>
            <div>
              <h4 class="text-xs font-bold text-white">Verifikasi Terenkripsi Pos Gerbang</h4>
              <p class="text-[11px] text-slate-400 mt-0.5">Pemindaian QR kriptografis real-time untuk menjamin otentisitas izin barang bawaan teknisi.</p>
            </div>
          </div>

          <div class="flex items-start gap-3 p-3 rounded-xl bg-slate-800/60 border border-slate-700/60">
            <div class="w-8 h-8 rounded-lg bg-inka-900/60 text-inka-400 flex items-center justify-center shrink-0 border border-inka-700/40">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
            <div>
              <h4 class="text-xs font-bold text-white">Alur Persetujuan & Audit Kepatuhan</h4>
              <p class="text-[11px] text-slate-400 mt-0.5">Sistem approval matriks otomatis dan rekam jejak audit yang tidak dapat diubah (immutable).</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Login Form Card (7 cols) -->
      <div class="lg:col-span-7 bg-white rounded-2xl border border-slate-200 shadow-2xl p-8">
        <div class="mb-6">
          <div class="flex items-center justify-between mb-3">
            <img src="/logo.png" alt="PT INKA (Persero)" class="h-7 w-auto object-contain" />
            <Badge variant="inka" size="sm">Akses Korporat</Badge>
          </div>
          <h2 class="text-xl font-bold text-slate-900">
            Masuk ke SIMAK
          </h2>
          <p class="text-xs text-slate-500 mt-1">
            Gunakan alamat email resmi kedinasan atau NIP terdaftar Anda
          </p>
        </div>

        <!-- Alert Error Message -->
        <div
          v-if="localError || authStore.error"
          class="mb-5 p-3 rounded-lg bg-rose-50 border border-rose-200 text-xs text-rose-700 flex items-start gap-2.5"
        >
          <svg class="w-4 h-4 text-rose-600 shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div class="flex-1">
            <p class="font-semibold">Autentikasi Gagal</p>
            <p>{{ localError || authStore.error }}</p>
          </div>
        </div>

        <form @submit.prevent="handleLogin" class="space-y-4">
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1.5">
              Alamat Email / NIP
            </label>
            <div class="relative">
              <input
                v-model="email"
                type="email"
                required
                autocomplete="email"
                placeholder="nama@inka.co.id"
                class="w-full px-3.5 py-2.5 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-inka-600 focus:bg-white transition-all pl-10"
              />
              <svg class="w-4 h-4 text-slate-400 absolute left-3.5 top-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.206" />
              </svg>
            </div>
          </div>

          <div>
            <div class="flex items-center justify-between mb-1.5">
              <label class="text-xs font-semibold text-slate-700">
                Kata Sandi
              </label>
              <span class="text-[11px] text-inka-700 hover:text-inka-900 cursor-pointer font-medium">
                Lupa sandi?
              </span>
            </div>
            <div class="relative">
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                required
                autocomplete="current-password"
                placeholder="••••••••"
                class="w-full px-3.5 py-2.5 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-inka-600 focus:bg-white transition-all pl-10 pr-10"
              />
              <svg class="w-4 h-4 text-slate-400 absolute left-3.5 top-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-3 top-2.5 text-slate-400 hover:text-slate-600 cursor-pointer"
                title="Tampilkan / Sembunyikan"
              >
                <svg v-if="!showPassword" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l18 18" />
                </svg>
              </button>
            </div>
          </div>

          <div class="pt-2">
            <Button
              type="submit"
              variant="primary"
              size="lg"
              :loading="authStore.loading"
              class="w-full text-sm font-bold shadow-md"
            >
              Masuk ke Sistem SIMAK
            </Button>
          </div>
        </form>

        <div class="mt-6 pt-5 border-t border-slate-100 flex items-center justify-between text-[11px] text-slate-400">
          <span>Sistem SIMAK PT INKA v1.0</span>
          <span>Keamanan Terenkripsi TLS / JWT</span>
        </div>
      </div>
    </div>
  </div>
</template>
