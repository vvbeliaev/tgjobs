import PocketBase from 'pocketbase';

/**
 * Script to update all jobs status to 'processed'
 * 
 * Usage:
 * PB_EMAIL=admin@example.com PB_PASSWORD=1234567890 node scripts/update-jobs.js
 */

const PB_URL = process.env.PUBLIC_PB_URL || 'http://127.0.0.1:8090';
const PB_EMAIL = process.env.PB_EMAIL;
const PB_PASSWORD = process.env.PB_PASSWORD;

if (!PB_EMAIL || !PB_PASSWORD) {
  console.error('Error: PB_EMAIL and PB_PASSWORD environment variables are required.');
  console.log('Usage: PB_EMAIL=your@email.com PB_PASSWORD=yourpassword node scripts/update-jobs.js');
  process.exit(1);
}

const pb = new PocketBase(PB_URL);

async function updateJobsStatus() {
  try {
    console.log(`Connecting to ${PB_URL}...`);
    
    // 1. Login as superuser
    // In PocketBase 0.23+, superusers are in the '_superusers' collection
    await pb.collection('_superusers').authWithPassword(PB_EMAIL, PB_PASSWORD);
    console.log('✅ Logged in successfully as superuser.');

    // 2. Fetch all jobs that are not 'processed'
    console.log('Fetching jobs...');
    const jobs = await pb.collection('jobs').getFullList({
      filter: 'status != "processed"',
    });

    if (jobs.length === 0) {
      console.log('✨ No jobs found that need updating.');
      return;
    }

    console.log(`Found ${jobs.length} jobs to update.`);

    // 3. Update each job
    // PocketBase SDK doesn't support batch updates yet, so we do them sequentially or in parallel chunks
    let count = 0;
    for (const job of jobs) {
      await pb.collection('jobs').update(job.id, {
        status: 'processed'
      });
      count++;
      if (count % 10 === 0 || count === jobs.length) {
        console.log(`Updated ${count}/${jobs.length} jobs...`);
      }
    }

    console.log('✅ All jobs updated to "processed" status.');
  } catch (error) {
    console.error('❌ Error:', error.message);
    if (error.data) {
      console.error('Details:', JSON.stringify(error.data, null, 2));
    }
    process.exit(1);
  }
}

updateJobsStatus();
