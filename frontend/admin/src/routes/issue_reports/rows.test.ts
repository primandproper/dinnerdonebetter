import { describe, expect, it } from 'vitest';
import { IssueReport, ReportStatus, reportStatusToJSON } from '@primandproper/platform-client/issuereports/v1';
import { toOperatorRows } from './rows';

function randomID(): string {
  return crypto.randomUUID();
}

describe('toOperatorRows', () => {
  it('carries each report with the account it was filed in', () => {
    const accountID = randomID();
    const report = IssueReport.create({
      id: randomID(),
      kind: randomID(),
      subjectType: randomID(),
      subjectId: randomID(),
      status: ReportStatus.REPORT_STATUS_OPEN,
      createdAt: new Date(),
    });

    const rows = toOperatorRows([{ scope: accountID, report }]);

    expect(rows).toEqual([
      {
        id: report.id,
        accountID,
        kind: report.kind,
        status: reportStatusToJSON(report.status),
        subject: `${report.subjectType}/${report.subjectId}`,
        createdAt: report.createdAt,
      },
    ]);
  });

  it('skips an entry with no report rather than rendering an empty row', () => {
    expect(toOperatorRows([{ scope: randomID(), report: undefined }])).toEqual([]);
  });

  it('leaves the subject empty for a report about nothing in particular', () => {
    const report = IssueReport.create({ id: randomID(), kind: randomID() });

    const [row] = toOperatorRows([{ scope: randomID(), report }]);

    expect(row.subject).toBe('');
  });
});
