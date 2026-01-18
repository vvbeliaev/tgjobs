<script lang="ts">
</script>

<div class="flex min-h-screen flex-col items-center justify-center overflow-hidden bg-base-100">
	<!-- Animated background gradient -->
	<div
		class="animate-gradient-shift absolute inset-0 bg-linear-to-br from-primary/5 via-accent/5 to-secondary/5"
	></div>

	<!-- Main content container -->
	<div class="relative z-10 flex flex-col items-center justify-center">
		<!-- Subtle loading dots -->
		<div class="flex gap-3">
			<div
				class="animate-dot-pulse h-2.5 w-2.5 rounded-full bg-accent/60 shadow-sm"
				style="animation-delay: 0s;"
			></div>
			<div
				class="animate-dot-pulse h-2.5 w-2.5 rounded-full bg-primary/60 shadow-sm"
				style="animation-delay: 0.2s;"
			></div>
			<div
				class="animate-dot-pulse h-2.5 w-2.5 rounded-full bg-secondary/60 shadow-sm"
				style="animation-delay: 0.4s;"
			></div>
		</div>
	</div>

	<!-- Bottom footer -->
	<div class="animate-fade-in-up absolute bottom-10 z-10">
		<a
			href="https://cogisoft.dev"
			target="_blank"
			rel="noopener noreferrer"
			class="group text-xs tracking-widest text-base-content/30 transition-all duration-500 hover:text-base-content/60"
		>
			made by <span class="font-medium">cogisoft.dev</span>
		</a>
	</div>
</div>

<style>
	@keyframes fade-in-up {
		from {
			opacity: 0;
			transform: translateY(12px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes dot-pulse {
		0%,
		100% {
			opacity: 0.3;
			transform: scale(0.95);
		}
		50% {
			opacity: 1;
			transform: scale(1.1);
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

	.animate-fade-in-up {
		animation: fade-in-up 1.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
	}

	.animate-dot-pulse {
		animation: dot-pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
	}

	.animate-gradient-shift {
		background-size: 200% 200%;
		animation: gradient-shift 12s ease infinite;
	}
</style>
