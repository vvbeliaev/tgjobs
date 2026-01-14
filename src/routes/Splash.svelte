<script lang="ts">
</script>

<div class="flex min-h-screen flex-col items-center justify-center overflow-hidden bg-base-100">
	<!-- Animated background gradient -->
	<div
		class="animate-gradient-shift absolute inset-0 bg-linear-to-br from-primary/5 via-accent/5 to-secondary/5"
	></div>

	<!-- Main heart container -->
	<div class="relative z-10 flex flex-col items-center justify-center">
		<!-- Pulsing heart with glow effect -->
		<div class="relative">
			<!-- Glow effect -->
			<div class="animate-heartbeat-glow absolute inset-0 rounded-full bg-accent/30 blur-2xl"></div>

			<!-- Heart icon wrapper with animation -->
			<div class="relative">
				<span class="text-center text-[4rem] font-bold text-primary">Job Hunter</span>
			</div>
		</div>

		<!-- Loading text with fade animation -->
		<h1
			class="animate-fade-in-up mt-10 text-center text-4xl font-semibold text-base-content drop-shadow-sm"
		>
			Loading app...
		</h1>

		<!-- Subtle loading dots with improved animation -->
		<div class="mt-6 flex gap-3">
			<div
				class="animate-dot-pulse h-3 w-3 rounded-full bg-accent shadow-sm"
				style="animation-delay: 0s;"
			></div>
			<div
				class="animate-dot-pulse h-3 w-3 rounded-full bg-primary shadow-sm"
				style="animation-delay: 0.2s;"
			></div>
			<div
				class="animate-dot-pulse h-3 w-3 rounded-full bg-secondary shadow-sm"
				style="animation-delay: 0.4s;"
			></div>
		</div>
	</div>
</div>

<style>
	@keyframes heartbeat-glow {
		0%,
		100% {
			opacity: 0.3;
			transform: scale(1);
		}
		50% {
			opacity: 0.5;
			transform: scale(1.2);
		}
	}

	@keyframes fade-in-up {
		from {
			opacity: 0;
			transform: translateY(24px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes dot-pulse {
		0%,
		100% {
			opacity: 0.4;
			transform: scale(0.9);
		}
		50% {
			opacity: 1;
			transform: scale(1.15);
		}
	}

	@keyframes gradient-shift {
		0%,
		100% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
	}

	.animate-heartbeat-glow {
		animation: heartbeat-glow 1.4s cubic-bezier(0.4, 0, 0.6, 1) infinite;
	}

	.animate-fade-in-up {
		animation: fade-in-up 0.9s cubic-bezier(0.16, 1, 0.3, 1) forwards;
	}

	.animate-dot-pulse {
		animation: dot-pulse 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
	}

	.animate-gradient-shift {
		background-size: 200% 200%;
		animation: gradient-shift 8s ease infinite;
	}
</style>
