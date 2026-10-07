import { describe, expect, it } from 'vitest';
import { IssueReport, ReportStatus, reportStatusToJSON } from '@primandproper/platform-client/issuereports/v1';
import { toOperatorRows } from './rows';

function randomID(): string {
  return crypto.randomUUID();
}

describe('toOperatorRows', () => {
  it('carries each report with who filed it', () => {
    const report = IssueReport.create({
      id: randomID(),
      reporter: randomID(),
      kind: randomID(),
      subjectType: randomID(),
      subjectId: randomID(),
      status: ReportStatus.REPORT_STATUS_OPEN,
      createdAt: new Date(),
    });

    const rows = toOperatorRows([report]);

    expect(rows).toEqual([
      {
        id: report.id,
        reporter: report.reporter,
        kind: report.kind,
        status: reportStatusToJSON(report.status),
        subject: `${report.subjectType}/${report.subjectId}`,
        createdAt: report.createdAt,
      },
    ]);
  });

  it('leaves the subject empty for a report about nothing in particular', () => {
    const report = IssueReport.create({ id: randomID(), kind: randomID() });

    const [row] = toOperatorRows([report]);

    expect(row.subject).toBe('');
  });
});
