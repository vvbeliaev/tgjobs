import { Collections, pb, type JobsResponse } from '$lib';

import { userJobsStore } from './user-jobs.svelte';

class JobsStore {
	private userId: string | null = null;

	jobs: JobsResponse[] = $state([]);
	search = $state('');
	filterRemote: boolean | null = $state(null);
	filterGrade = $state('');
	showArchived = $state(false);

	currentPage = $state(1);
	pageSize = $state(10);

	filteredJobs = $derived.by(() => {
		let result = this.jobs || [];

		if (!this.showArchived) {
			result = result.filter((j) => !userJobsStore.isArchived(j.id));
		} else {
			result = result.filter((j) => userJobsStore.isArchived(j.id));
		}

		if (this.search) {
			const s = this.search.toLowerCase();
			result = result.filter((j) => {
				return (
					j.title.toLowerCase().includes(s) ||
					j.company?.toLowerCase().includes(s) ||
					j.description?.toLowerCase().includes(s)
				);
			});
		}

		if (this.filterRemote !== null) {
			result = result.filter((j) => j.isRemote === this.filterRemote);
		}

		if (this.filterGrade) {
			result = result.filter((j) =>
				j.grade?.toLowerCase().includes(this.filterGrade.toLowerCase())
			);
		}

		return result;
	});

	paginatedJobs = $derived.by(() => {
		const start = (this.currentPage - 1) * this.pageSize;
		const end = start + this.pageSize;
		return this.filteredJobs.slice(start, end);
	});

	totalPages = $derived(Math.ceil(this.filteredJobs.length / this.pageSize));

	constructor() {
		$effect.root(() => {
			$effect(() => {		
				void [this.search, this.filterRemote, this.filterGrade, this.showArchived];
				this.currentPage = 1;
			});
		});
	}

	async load(userId: string) {
		this.userId = userId;

		// Load jobs directly
		const jobs = await pb.collection(Collections.Jobs).getFullList<JobsResponse>({
			filter: 'status = "processed"',
			sort: '-created'
		});

		this.jobs = jobs;

		return this.jobs;
	}

	set(jobs: JobsResponse[]) {
		this.jobs = jobs;
	}

	async subscribe() {
		if (!this.userId) return;

		await userJobsStore.subscribe();

		return pb.collection(Collections.Jobs).subscribe<JobsResponse>('*', async (e) => {
			switch (e.action) {
				case 'create': {
					break;
				}
				case 'update': {
					const exists = this.jobs.find((j) => j.id === e.record.id);
					if (exists) {
						this.jobs = this.jobs.map((j) => (j.id === e.record.id ? e.record : j));
					} else {
						this.jobs.unshift(e.record);
					}
					break;
				}
				case 'delete':
					this.jobs = this.jobs.filter((j) => j.id !== e.record.id);
					break;
			}
		}, 
		{
			filter: `status = "processed" && userId = "${this.userId}"`
		});
	}

	unsubscribe() {
		pb.collection(Collections.Jobs).unsubscribe('*');
		userJobsStore.unsubscribe();
	}

	clear() {
		this.jobs = [];
		this.userId = null;
		this.currentPage = 1;
		userJobsStore.clear();
	}
}

export const jobsStore = new JobsStore();
