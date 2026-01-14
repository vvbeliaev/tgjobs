<script lang="ts">
	import { Crosshair } from 'lucide-svelte';
	import type { ClassValue } from 'svelte/elements';

	interface Props {
		size?: number;
		class?: ClassValue;
		text?: string;
	}

	const { size = 24, class: className = '', text = 'Job Hunter' }: Props = $props();
</script>

<div
	class="flex size-9 items-center justify-center rounded-xl bg-primary text-white dark:text-black"
>
	<Crosshair {size} />
</div>
<span class="font-display text-xl font-bold tracking-tight text-nowrap">{text}</span>
