package gocontext

import (
	"context" // Mengelola context untuk sinyal cancel, timeout, dan value
	"fmt"     // Mencetak teks / output ke terminal
	"runtime" // Memantau info runtime Go, contohnya menghitung jumlah goroutine aktif
	"testing" // Framework unit test bawaan Go
	"time"    // Mengelola waktu, durasi, dan jeda (sleep)
)

func TestContext(t *testing.T) {
	// Membuat context kosong paling dasar (akar / root dari pohon context)
	background := context.Background()
	// Menampilkan context background ke terminal
	fmt.Println(background)

	// Membuat context kosong penanda jika belum tahu context apa yang ingin digunakan
	todo := context.TODO()
	// Menampilkan context todo ke terminal
	fmt.Println(todo)
}

func TestContextWithValue(t *testing.T) {
	// Membuat context induk (root) tanpa nilai
	contextA := context.Background()

	// Membuat context B anak dari A, membawa key "b" dan value "B"
	contextB := context.WithValue(contextA, "b", "B")
	// Membuat context C anak dari A, membawa key "c" dan value "C"
	contextC := context.WithValue(contextA, "c", "C")

	// Membuat context D anak dari B, membawa key "d" dan value "D" (mewarisi nilai B)
	contextD := context.WithValue(contextB, "d", "D")
	// Membuat context E anak dari B, membawa key "e" dan value "E" (mewarisi nilai B)
	contextE := context.WithValue(contextB, "e", "E")

	// Membuat context F anak dari C, membawa key "f" dan value "F" (mewarisi nilai C)
	contextF := context.WithValue(contextC, "f", "F")

	// Menampilkan seluruh context untuk melihat hierarki struktur nilainya
	fmt.Println(contextA)
	fmt.Println(contextB)
	fmt.Println(contextC)
	fmt.Println(contextD)
	fmt.Println(contextE)
	fmt.Println(contextF)
}

func CreateCounter(ctx context.Context) chan int {
	// Membuat unbuffered channel sebagai pipa penyalur angka keluar
	destination := make(chan int)

	// Menjalankan proses pengiriman angka di goroutine baru (background worker)
	go func() {
		// Menutup channel secara otomatis saat goroutine ini selesai / return
		defer close(destination)

		// Variabel penghitung angka awal dimulai dari 1
		counter := 1

		// Perulangan tanpa henti (infinite loop)
		for {
			select {
			// Mengecek apakah ada sinyal batal atau waktu habis dari context
			case <-ctx.Done():
				// Langsung keluar dari goroutine agar tidak terjadi kebocoran (goroutine leak)
				return
			// Jika belum ada sinyal cancel dari context, jalankan blok default
			default:
				destination <- counter      // Mengirim angka saat ini ke channel destination
				counter++                   // Menaikkan nilai counter sebesar 1
				time.Sleep(1 * time.Second) // Simulasi jeda / proses yang memakan waktu (1 detik)
			}
		}
	}()

	// Mengembalikan channel ke pemanggil agar bisa dibaca dari luar fungsi
	return destination
}

func TestContextWithCancel(t *testing.T) {
	// Menampilkan jumlah goroutine yang sedang aktif sebelum proses dimulai (kondisi awal)
	fmt.Println("Total Goroutine:", runtime.NumGoroutine())

	// Membuat context induk kosong
	parent := context.Background()

	// Membuat context baru turunan parent dan fungsi saklar (cancel) untuk membatalkannya manual
	ctx, cancel := context.WithCancel(parent)

	// Menjalankan goroutine counter di background dengan menyerahkan context
	destination := CreateCounter(ctx)

	// Menampilkan jumlah goroutine (akan bertambah 1 karena CreateCounter melahirkan goroutine baru)
	fmt.Println("Total Goroutine:", runtime.NumGoroutine())

	// Membaca angka satu per satu dari channel destination
	for n := range destination {
		fmt.Println("Counter", n) // Mencetak angka yang diterima ke terminal
		if n == 10 {
			break // Berhenti membaca channel setelah angka mencapai 10
		}
	}

	// Mengirim sinyal cancel manual ke context agar goroutine CreateCounter berhenti bekerja
	cancel()

	// Memberi jeda waktu 2 detik agar goroutine di background sempat merespons dan mati
	time.Sleep(2 * time.Second)

	// Menampilkan jumlah goroutine akhir (harus kembali ke jumlah awal, membuktikan tidak ada leak)
	fmt.Println("Total goroutine:", runtime.NumGoroutine())
}

func TestContextWithTimeout(t *testing.T) {
	// Menampilkan jumlah goroutine yang sedang aktif di awal
	fmt.Println("Total Goroutine:", runtime.NumGoroutine())

	// Membuat context induk kosong
	parent := context.Background()

	// Membuat context yang otomatis mengirim sinyal batal setelah durasi 5 detik habis
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)

	// Praktik baik (best practice): selalu panggil cancel() saat fungsi selesai agar resource timer dibersihkan
	defer cancel()

	// Menjalankan goroutine counter di background dengan membawa context ber-timeout
	destination := CreateCounter(ctx)

	// Menampilkan jumlah goroutine (bertambah 1 karena ada goroutine baru berjalan)
	fmt.Println("Total Goroutine:", runtime.NumGoroutine())

	// Membaca data tanpa henti secara otomatis; loop selesai sendiri saat channel ditutup di detik ke-5
	for n := range destination {
		fmt.Println("Counter", n) // Mencetak angka yang diterima setiap 1 detik
	}

	// Memanggil cancel untuk memastikan pembatalan dan pembersihan context
	cancel()

	// Memberi jeda 2 detik agar runtime Go menyelesaikan pembersihan goroutine
	time.Sleep(2 * time.Second)

	// Menampilkan jumlah goroutine akhir (kembali ke jumlah awal, membuktikan tidak ada kebocoran)
	fmt.Println("Total goroutine:", runtime.NumGoroutine())
}