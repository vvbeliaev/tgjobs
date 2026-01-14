import { Collections, pb, type JobsResponse } from '$lib';

class JobsStore {
	jobs: JobsResponse[] = $state([]);
	search = $state('');
	filterRemote: boolean | null = $state(null);
	filterGrade = $state('');
	showArchived = $state(false);

	filteredJobs = $derived.by(() => {
		let result = this.jobs || [];

		if (!this.showArchived) {
			result = result.filter((j) => !j.archived);
		} else {
			result = result.filter((j) => !!j.archived);
		}

		if (this.search) {
			const s = this.search.toLowerCase();
			result = result.filter(
				(j) =>
					j.title.toLowerCase().includes(s) ||
					j.company?.toLowerCase().includes(s) ||
					j.description?.toLowerCase().includes(s)
			);
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

	async load() {
		const jobs = await pb.collection(Collections.Jobs).getFullList({
			sort: '-created'
		});

		this.jobs = jobs;

		return jobs;
	}

	async toggleArchive(jobId: string) {
		const job = this.jobs.find((j) => j.id === jobId);
		if (!job) return;

		const archived = job.archived ? null : new Date().toISOString();
		await pb.collection(Collections.Jobs).update(jobId, { archived });
	}

	set(jobs: JobsResponse[]) {
		this.jobs = jobs;
	}

	async subscribe() {
		return pb.collection(Collections.Jobs).subscribe('*', (e) => {
			switch (e.action) {
				case 'create':
					this.jobs.unshift(e.record);
					break;
				case 'update':
					this.jobs = this.jobs.map((job) => (job.id === e.record.id ? e.record : job));
					break;
				case 'delete':
					this.jobs = this.jobs.filter((job) => job.id !== e.record.id);
					break;
			}
		});
	}

	unsubscribe() {
		pb.collection(Collections.Jobs).unsubscribe('*');
	}
}

export const jobsStore = new JobsStore();
